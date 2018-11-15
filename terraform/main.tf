/**
 * Creates the infrastructure to run an instance of Railgun
 *
 *
 *
 * Creates the following resources;
 *
 * * KMS - Key for Chamber
 * * ACM Certificate
 * * VPC - Virtual Private Cloud
 * * ALB - Aplication Load Balancer
 * * ECS/Fargate Cluster & Service
 * * S3 Bucket for Data
 * * S3 Bucket for Logs
 *
*/

#
# KMS - Key Management Service
#

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "key_policy_chamber" {

  statement {
    sid = "Enable IAM User Permissions"

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    
    actions = ["kms:*"]

    resources = ["*"]
  }
  
  statement {
    sid = "Allow access for Key Administrators"

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/admin"]
    }
    
    actions = [
      "kms:Create*",
      "kms:Describe*",
      "kms:Enable*",
      "kms:List*",
      "kms:Put*",
      "kms:Update*",
      "kms:Revoke*",
      "kms:Disable*",
      "kms:Get*",
      "kms:Delete*",
      "kms:TagResource",
      "kms:UntagResource",
      "kms:ScheduleKeyDeletion",
      "kms:CancelKeyDeletion"
    ]

    resources = ["*"]

  }
  
  statement {
    sid = "Allow use of the key"

    principals {
      type        = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ecs-task-role-railgun-${var.environment}",
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/admin",
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/engineer"
      ]
    }
    
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]

    resources = ["*"]

  }
  
  statement {
    sid = "Allow attachment of persistent resources"

    principals {
      type        = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/ecs-task-role-railgun-${var.environment}",
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/admin",
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/engineer"
      ]
    }
    
    actions = [
      "kms:CreateGrant",
      "kms:ListGrants",
      "kms:RevokeGrant"
    ]

    resources = ["*"]
    
    condition {
      test     = "Bool"
      variable = "kms:GrantIsForAWSResource"

      values = ["true"]
    }

  }

}

resource "aws_kms_key" "chamber" {
  description             = "Key for use with chamber. https://github.com/segmentio/chamber"
  deletion_window_in_days = 7
  policy = "${data.aws_iam_policy_document.key_policy_chamber.json}"
}

resource "aws_kms_alias" "chamber" {
  name          = "alias/chamber"
  target_key_id = "${aws_kms_key.chamber.key_id}"
}

#
# ACM - Certificate for Domain Name
#

data "aws_route53_zone" "main" {
  name = "${var.zone_name}"
  private_zone = false
}

resource "aws_acm_certificate" "main" {
  domain_name       = "${var.domain_name}"
  validation_method = "DNS"

  tags = {
    Name        = "${var.domain_name}"
    Environment = "${var.environment}"
    Automation  = "Terraform"
  }
}

resource "aws_route53_record" "cert_validation" {
  zone_id = "${data.aws_route53_zone.main.id}"
  name    = "${aws_acm_certificate.main.domain_validation_options.0.resource_record_name}"
  type    = "${aws_acm_certificate.main.domain_validation_options.0.resource_record_type}"
  records = ["${aws_acm_certificate.main.domain_validation_options.0.resource_record_value}"]
  ttl     = "60"
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn         = "${aws_acm_certificate.main.arn}"
  validation_record_fqdns = ["${aws_route53_record.cert_validation.fqdn}"]
}

#
# VPC - Virtual Private Cloud for Service
#

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "railgun-${var.environment}"
  cidr = "10.0.0.0/16"

  azs             = ["us-west-2a", "us-west-2b", "us-west-2c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = true
  one_nat_gateway_per_az = false

  tags = {
    Automation = "Terraform"
    Environment = "${var.environment}"
    Application = "Railgun"
  }
}

#
# ALB - Application Load Balancer to Front for Railgun
#

module "alb" {
  source = "trussworks/alb-web-containers/aws"
  version = "2.2.0"

  name           = "railgun"
  environment    = "${var.environment}"
  logs_s3_bucket = "${var.s3_bucket_logs}"

  alb_vpc_id                  = "${module.vpc.vpc_id}"
  alb_subnet_ids              = "${module.vpc.public_subnets}"
  alb_default_certificate_arn = "${aws_acm_certificate.main.arn}"

  container_port    = "${var.container_port}"
  container_protocol = "${var.container_protocol}"
  health_check_path = "${var.health_check_path}"
}

resource "aws_route53_record" "main" {
  zone_id = "${data.aws_route53_zone.main.id}"
  name    = "${var.domain_name}"
  type = "A"

  alias {
    name                   = "${module.alb.alb_dns_name}"
    zone_id                = "${module.alb.alb_zone_id}"
    evaluate_target_health = true
  }
}

#
# ECS - Fargate ECS Service for Railgun
#

resource "aws_ecs_cluster" "main" {
  name = "railgun-cluster"
}

data "aws_region" "current" {}

locals {

  container_definitions = <<EOF
[
  {
    "name": "${var.container_name}",
    "image": "${aws_ecr_repository.main.repository_url}:${var.image_tag}",
    "cpu": 0,
    "memory": ${var.task_memory},
    "essential": true,
    "portMappings": [
      {
        "containerPort": ${var.container_port},
        "hostPort": ${var.container_port},
        "protocol": "tcp"
      }
    ],
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "/ecs/${var.environment}/${var.container_name}",
        "awslogs-region": "${data.aws_region.current.name}",
        "awslogs-stream-prefix": "ecs"
      }
    },
    "entryPoint": [
      "/chamber",
      "exec",
      "railgun-${var.environment}",
      "--",
      "/railgun"
    ],
    "command": [
      "serve",
      "--verbose"
    ],
    "environment": [
      {
        "name": "AWS_DEFAULT_REGION",
        "value": "us-west-2"
      },
      {
        "name": "HTTP_ADDRESS",
        "value": "0.0.0.0:${var.container_port}"
      },
      {
        "name": "HTTP_MIDDLEWARE_GZIP",
        "value": "1"
      },
      {
        "name": "HTTP_LOCATION",
        "value": "https://railgun.spatialcurrent.io/"
      },
      {
        "name": "HTTP_SCHEMES",
        "value": "https"
      }
    ],
    "mountPoints": [],
    "volumesFrom": []
  }
]
EOF

}

module "ecs_service" {
  source = "trussworks/ecs-service/aws"

  name        = "railgun"
  environment = "${var.environment}"

  ecs_cluster_arn               = "${aws_ecs_cluster.main.id}"
  ecs_vpc_id                    = "${module.vpc.vpc_id}"
  ecs_subnet_ids                = "${module.vpc.private_subnets}"
  ecs_use_fargate = true

  tasks_desired_count           = 1
  tasks_minimum_healthy_percent = 50
  tasks_maximum_percent         = 200

  associate_alb      = 1
  alb_security_group = "${module.alb.alb_security_group_id}"
  lb_target_group   = "${module.alb.alb_target_group_id}"
  
  target_container_name = "${var.container_name}"
  container_port = "${var.container_port}"
  
  container_definitions = "${local.container_definitions}"
  
  fargate_task_cpu = "${var.task_cpu}"
  fargate_task_memory = "${var.task_memory}"
  
}

#
# S3 - Bucket for Logs
#

# https://stackoverflow.com/questions/43366038/terraform-elb-s3-permissions-issue
data "aws_elb_service_account" "main" {}

data "aws_iam_policy_document" "bucket_policy_logs" {

  statement {
    sid = "ensure-private-read-write"

    actions = [
      "s3:PutObject",
      "s3:PutObjectAcl"
    ]

    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    resources = ["arn:aws:s3:::${var.s3_bucket_logs}/*"]

    condition {
      test     = "StringEquals"
      variable = "s3:x-amz-acl"

      values = [
        "public-read",
        "public-read-write"
      ]
    }

  }

  statement {
    sid = "allow-elb-account"

    actions = [
      "s3:PutObject"
    ]

    principals {
      type        = "AWS"
      identifiers = ["${data.aws_elb_service_account.main.arn}"]
    }

    resources = ["arn:aws:s3:::${var.s3_bucket_logs}/*"]

  }

  statement {
    sid = "allow-alb-delivery-check"

    actions = [
      "s3:GetBucketAcl"
    ]

    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }

    resources = ["arn:aws:s3:::${var.s3_bucket_logs}"]

  }

  statement {
    sid = "allow-alb-delivery-write"

    actions = [
      "s3:PutObject"
    ]

    principals {
      type        = "Service"
      identifiers = ["delivery.logs.amazonaws.com"]
    }

    resources = ["arn:aws:s3:::${var.s3_bucket_logs}/*"]

    condition {
      test     = "StringEquals"
      variable = "s3:x-amz-acl"

      values = [
        "bucket-owner-full-control"
      ]
    }

  }

}

resource "aws_s3_bucket" "logs" {
  bucket = "${var.s3_bucket_logs}"
  acl    = "private"
  policy = "${data.aws_iam_policy_document.bucket_policy_logs.json}"
  
  tags = {
    Automation = "Terraform"
    Environment = "${var.environment}"
    Application = "Railgun"
  }

  versioning {
    enabled = true
  }

  lifecycle_rule {
    enabled = true

    abort_incomplete_multipart_upload_days = 14

    expiration {
      expired_object_delete_marker = true
    }

    noncurrent_version_transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    noncurrent_version_expiration {
      days = 365
    }
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}

#
# S3 - Bucket for Catalog & Data
#

data "aws_iam_policy_document" "bucket_policy_data" {

  statement {
    sid = "ensure-private-read-write"

    actions = [
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]

    effect = "Deny"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    resources = ["arn:aws:s3:::${var.s3_bucket_data}/*"]

    condition {
      test     = "StringEquals"
      variable = "s3:x-amz-acl"

      values = [
        "public-read",
        "public-read-write",
      ]
    }
  }
}

resource "aws_s3_bucket" "data" {
  bucket = "${var.s3_bucket_data}"
  acl    = "private"
  policy = "${data.aws_iam_policy_document.bucket_policy_data.json}"
  
  tags = {
    Automation = "Terraform"
    Environment = "${var.environment}"
    Application = "Railgun"
  }

  versioning {
    enabled = true
  }

  lifecycle_rule {
    enabled = true

    abort_incomplete_multipart_upload_days = 14

    expiration {
      expired_object_delete_marker = true
    }

    noncurrent_version_transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    noncurrent_version_expiration {
      days = 365
    }
  }

  logging {
    target_bucket = "${var.s3_bucket_logs}"
    target_prefix = "logs/s3/${var.s3_bucket_data}/"
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}

#
# ECR - Docker Respository
#

resource "aws_ecr_repository" "main" {
  name = "railgun"
}

#
# IAM - System Security Manager Parameter Store Access
#

data "aws_iam_policy_document" "ssm" {
  statement {

    actions = [
      "ssm:PutParameter",
      "ssm:DeleteParameter",
      "ssm:DescribeParameters",
      "ssm:GetParameterHistory",
      "ssm:GetParametersByPath",
      "ssm:GetParameters",
      "ssm:GetParameter",
      "ssm:DeleteParameters"
    ]

    resources = [
      "*"
    ]
  }
}

resource "aws_iam_policy" "ssm" {
  name   = "ssm-allow-params"
  path   = "/"
  policy = "${data.aws_iam_policy_document.ssm.json}"
}

resource "aws_iam_role_policy_attachment" "ssm" {
  role       = "${module.ecs_service.task_role_name}"
  policy_arn = "${aws_iam_policy.ssm.arn}"
}

#
# IAM - Cloudwatch Access
#

data "aws_iam_policy_document" "cloudwatch" {
  statement {

    actions = [
      "logs:CreateLogStream",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
      "logs:CreateLogGroup",
      "logs:PutLogEvents"
    ]

    resources = [
      "*"
    ]
  }
}

resource "aws_iam_policy" "cloudwatch" {
  name   = "cloudwatch-allow-write"
  path   = "/"
  policy = "${data.aws_iam_policy_document.cloudwatch.json}"
}

resource "aws_iam_role_policy_attachment" "cloudwatch" {
  role       = "ecs-task-execution-role-railgun-${var.environment}"
  policy_arn = "${aws_iam_policy.cloudwatch.arn}"
}

#
# IAM - ECR Access
#

data "aws_iam_policy_document" "ecr" {
  statement {

    actions = [
      "ecr:GetAuthorizationToken"
    ]

    resources = [
      "*"
    ]
  }
  
  statement {
  
    actions = [
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "ecr:DescribeImages",
      "ecr:ListImages",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetRepositoryPolicy"
    ]
  
    resources = [
      "${aws_ecr_repository.main.arn}"
    ]
  }

}

resource "aws_iam_policy" "ecr" {
  name   = "ecr-allow-read"
  path   = "/"
  policy = "${data.aws_iam_policy_document.ecr.json}"
}

resource "aws_iam_role_policy_attachment" "ecr" {
  role       = "ecs-task-execution-role-railgun-${var.environment}"
  policy_arn = "${aws_iam_policy.ecr.arn}"
}

#
# IAM - Access to S3 Catalog & Data
#

data "aws_iam_policy_document" "s3" {
  statement {

    actions = [
      "s3:ListBucket"
    ]

    resources = [
      "${aws_s3_bucket.data.arn}"
    ]
  }
  
  statement {
  
    actions = [
      "s3:PutObject",
      "s3:GetObjectAcl",
      "s3:GetObject",
      "s3:GetObjectVersionAcl",
      "s3:GetObjectVersion"
    ]
  
    resources = [
      "${aws_s3_bucket.data.arn}/*"
    ]
  }

}

resource "aws_iam_policy" "s3" {
  name   = "s3-allow-write"
  path   = "/"
  policy = "${data.aws_iam_policy_document.s3.json}"
}

resource "aws_iam_role_policy_attachment" "s3" {
  role       = "${module.ecs_service.task_role_name}"
  policy_arn = "${aws_iam_policy.s3.arn}"
}