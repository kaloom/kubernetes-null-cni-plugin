#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$(readlink -f "../${BASH_SOURCE[0]}")")"

[ -x bin/null ] || (echo "please build the null cni-plugin first by running ./build.sh"; exit 1)

. gradle.properties

docker build . -t kaloom/null:$version
