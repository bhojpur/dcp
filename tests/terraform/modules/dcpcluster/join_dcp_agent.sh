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

# This script is used to join one or more nodes as agents

mkdir -p /etc/bhojpur/dcp
cat <<EOF >>/etc/bhojpur/dcp/config.yaml
server: https://${4}:6443
token:  "${5}"
EOF

if [[ ! -z "$7" ]] && [[ "$7" == *":"* ]]
then
   echo -e "$7" >> /etc/bhojpur/dcp/config.yaml
   cat /etc/bhojpur/dcp/config.yaml
fi

if [ ${1} = "rhel" ]
then
    subscription-manager register --auto-attach --username=${8} --password=${9}
    subscription-manager repos --enable=rhel-7-server-extras-rpms
fi

export "${2}"="${3}"
  if [[ "$3" == *"v1.18"* ]] || [["$3" == *"v1.17"* ]] && [[ -n "$7" ]]
then
  echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | sh -s - agent --node-external-ip=${6} $7" >/tmp/agent_cmd
curl -sfL https://get.bhojpur.net/dcp/install.sh | sh -s - agent --node-external-ip=${6} ${7}
  else

echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | sh -s - agent --node-external-ip=${6}" >/tmp/agent_cmd
curl -sfL https://get.bhojpur.net/dcp/install.sh | sh -s - agent --node-external-ip=${6}
fi