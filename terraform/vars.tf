variable "environment" {
  type = "string"
  default = "prod"
}

variable "zone_name" {
  type = "string"
  default = "spatialcurrent.io"
}

variable "domain_name" {
  type = "string"
  default = "railgun.spatialcurrent.io"
}

variable "s3_bucket_data" {
  type = "string"
  default = "spatialcurrent-data-us-west-2"
}

variable "s3_bucket_logs" {
  type = "string"
  default = "spatialcurrent-logs-us-west-2"
}

variable "health_check_path" {
  type = "string"
  default = "/"
}

variable "container_protocol" {
  type = "string"
  default = "HTTP"
}

variable "container_port" {
  type = "string"
  default = "8000"
}

variable "container_name" {
  type = "string"
  default = "railgun"
}

variable "image_tag" {
  type = "string"
  default = "git-13a201151b75cb4efdab74d1c1f63913257fec3d"
}

variable "task_cpu" {
  type = "string"
  default = "2048"
}

variable "task_memory" {
  type = "string"
  default = "4096"
}