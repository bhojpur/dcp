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

mkdir -p /etc/bhojpur/dcp
cat << EOF >/etc/bhojpur/dcp/config.yaml
write-kubeconfig-mode: "0644"
tls-san:
  - ${2}
EOF

if [[ -n "$8" ]] && [[ "$8" == *":"* ]]
then
   echo "$"
   echo -e "$8" >> /etc/bhojpur/dcp/config.yaml
   cat /etc/bhojpur/dcp/config.yaml
fi

if [ "${1}" = "rhel" ]
then
   subscription-manager register --auto-attach --username="${9}" --password="${10}"
   subscription-manager repos --enable=rhel-7-server-extras-rpms
fi

export "${3}"="${4}"

if [ "${5}" = "etcd" ]
then
   echo "CLUSTER TYPE  is etcd"
   if [[ "$4" == *"v1.18"* ]] || [["$4" == *"v1.17"* ]] && [[ -n "$8" ]]
   then
       echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --cluster-init --node-external-ip=${6} $8" >/tmp/master_cmd
       curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --cluster-init --node-external-ip="${6}" "$8"
   else
       echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --cluster-init --node-external-ip=${6}" >/tmp/master_cmd
       curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --cluster-init --node-external-ip="${6}"
   fi
else
   echo "CLUSTER TYPE is external db"
   echo "$8"
   if [[ "$4" == *"v1.18"* ]] || [[ "$4" == *"v1.17"* ]] && [[ -n "$8" ]]
   then
       echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | sh -s - server --node-external-ip=${6} --datastore-endpoint=\"${7}\" $8"  >/tmp/master_cmd
       curl -sfL https://get.bhojpur.net/dcp/install.sh | sh -s - server --node-external-ip="${6}" --datastore-endpoint="${7}" "$8"
   else
       echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | sh -s - server --node-external-ip=${6}  --datastore-endpoint=\"${7}\" "  >/tmp/master_cmd
       curl -sfL https://get.bhojpur.net/dcp/install.sh | sh -s - server --node-external-ip="${6}" --datastore-endpoint="${7}"
   fi
fi

export PATH=$PATH:/usr/local/bin
timeElapsed=0
while ! $(kubectl get nodes >/dev/null 2>&1) && [[ $timeElapsed -lt 300 ]]
do
   sleep 5
   timeElapsed=$(expr $timeElapsed + 5)
done

IFS=$'\n'
timeElapsed=0
sleep 10
while [[ $timeElapsed -lt 420 ]]
do
   notready=false
   for rec in $(kubectl get nodes)
   do
      if [[ "$rec" == *"NotReady"* ]]
      then
         notready=true
      fi
  done
  if [[ $notready == false ]]
  then
     break
  fi
  sleep 20
  timeElapsed=$(expr $timeElapsed + 20)
done

IFS=$'\n'
timeElapsed=0
while [[ $timeElapsed -lt 420 ]]
do
   helmPodsNR=false
   systemPodsNR=false
   for rec in $(kubectl get pods -A --no-headers)
   do
      if [[ "$rec" == *"helm-install"* ]] && [[ "$rec" != *"Completed"* ]]
      then
         helmPodsNR=true
      elif [[ "$rec" != *"helm-install"* ]] && [[ "$rec" != *"Running"* ]]
      then
         systemPodsNR=true
      else
         echo ""
      fi
   done

   if [[ $systemPodsNR == false ]] && [[ $helmPodsNR == false ]]
   then
      break
   fi
   sleep 20
   timeElapsed=$(expr $timeElapsed + 20)
done
cat /etc/bhojpur/dcp/config.yaml> /tmp/joinflags
cat /var/lib/bhojpur/dcp/server/node-token >/tmp/nodetoken
cat /etc/bhojpur/dcp/dcp.yaml >/tmp/config