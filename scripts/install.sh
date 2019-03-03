#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
set -eu
cd $DIR/..
pkgs=$(go list ./... | grep cmd | grep -v '.js')
echo "******************"
echo "Installing programs"
for pkg in "${pkgs[@]}"; do
  echo "Installing $(basename $pkg) from $pkg"
  go install $pkg
done
