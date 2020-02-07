#!/bin/bash

set -euo pipefail

cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"

# Add branch/commit/date into binary
set +e
git describe --tags --abbrev=0 > /dev/null 2>&1
if [ "$?" != "0" ]; then
    BRANCH="master"
else
    BRANCH=$(git describe --tags --abbrev=0)
fi

set -e
DATE=$(date --utc "+%F_%H:%m:%S_+0000")
COMMIT=$(git rev-parse --verify --short HEAD)
LDFLAGS="-X main.branch=${BRANCH:-master} -X main.commit=${COMMIT} -X main.date=${DATE}"

repo_path="github.com/kaloom/kubernetes-null-cni-plugin"
exec_name="null"
org_path=$(echo $repo_path | cut -d/ -f 1-2)

export GOBIN=${PWD}/bin
export GO111MODULE=on

echo "Building $exec_name cni-plugin"
go install -ldflags "${LDFLAGS}" "$@" ${repo_path}/${exec_name}
