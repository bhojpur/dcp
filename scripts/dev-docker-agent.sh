#!/bin/bash

# Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

set -e

cd $(dirname $0)/..

. ./scripts/setup-bhojpur-path.sh

IP=$(ip addr show dev docker0 | grep -w inet | awk '{print $2}' | cut -f1 -d/)
docker run \
    --read-only \
    --tmpfs /run \
    --tmpfs /var/run \
    --tmpfs /tmp \
    -v /lib/modules:/lib/modules:ro \
    -v /lib/firmware:/lib/firmware:ro \
    -v /etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt:ro \
    -v $(pwd)/bin:/usr/bin \
    -v /var/log \
    -v /var/lib/kubelet \
    -v /var/lib/bhojpur/dcp \
    -v /var/lib/cni \
    -v /usr/lib/x86_64-linux-gnu/libsqlite3.so.0:/usr/lib/x86_64-linux-gnu/libsqlite3.so.0:ro \
    --privileged \
    ubuntu:18.04 /usr/bin/dcp-agent agent -t $(<${BHOJPUR_PATH}/dcp/server/node-token) -s https://${IP}:6443