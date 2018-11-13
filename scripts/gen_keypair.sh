#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
mkdir -p $DIR/../secret
ssh-keygen -t rsa -b 4096 -m PEM -f $DIR/../secret/keypair
openssl rsa -in $DIR/../secret/keypair -pubout -outform PEM -out $DIR/../secret/keypair.pub