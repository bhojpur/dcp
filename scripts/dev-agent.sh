#!/bin/bash
set -e

cd $(dirname $0)/..

. ./scripts/setup-bhojpur-path.sh

GO=${GO-go}

# Prime sudo
sudo echo Compiling

if [ ! -e bin/containerd ]; then
    ./scripts/build
    ./scripts/package
else
    rm -f ./bin/dcp-agent
    "${GO}" build -tags "apparmor seccomp" -o ./bin/dcp-agent ./cmd/cloud/agent/main.go
fi

echo Starting Bhojpur DCP agent
sudo env "PATH=$(pwd)/bin:$PATH" ./bin/dcp-agent --debug agent -s https://localhost:6443 -t $(<${BHOJPUR_PATH}/dcp/server/node-token) "$@"