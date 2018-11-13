#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cat $DIR/../secret/keypair.pub | CHAMBER_KMS_KEY_ALIAS=chamber chamber write railgun-prod jwt-public-key -
cat $DIR/../secret/keypair | CHAMBER_KMS_KEY_ALIAS=chamber chamber write railgun-prod jwt-private-key -