#!/bin/bash
# This script is used to join one or more nodes as masters

mkdir -p /etc/bhojpur/dcp
cat <<EOF >>/etc/bhojpur/dcp/config.yaml
write-kubeconfig-mode: "0644"
tls-san:
  - ${2}
EOF

if [[ -n "${10}" ]] && [[ "${10}" == *":"* ]]
then
   echo -e "${10}" >> /etc/bhojpur/dcp/config.yaml
   cat /etc/bhojpur/dcp/config.yaml
fi

if [ "${1}" = "rhel" ]
then
   subscription-manager register --auto-attach --username="${11}" --password="${12}"
   subscription-manager repos --enable=rhel-7-server-extras-rpms
fi

export "${3}"="${4}"

if [ "${5}" = "etcd" ]
then
   if [[ "$4" == *"v1.18"* ]] || [["$4" == *"v1.17"* ]] && [[ -n "$10" ]]
   then
      echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --server https://\"${7}\":6443 --token \"${8}\" --node-external-ip=\"${6}\" ${10}" >/tmp/master_cmd
      curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --server https://"${7}":6443 --token "${8}" --node-external-ip="${6} ${10}"
   else
        echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --server https://\"${7}\":6443 --token \"${8}\" --node-external-ip=\"${6}\"" >/tmp/master_cmd
       curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --server https://"${7}":6443 --token "${8}" --node-external-ip="${6}"
   fi
else
  if [[ "$4" == *"v1.18"* ]] || [["$4" == *"v1.17"* ]] && [[ -n "$10" ]]
  then
      echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --node-external-ip=\"${6}\" --datastore-endpoint=\"${9}\" ${10}" >/tmp/master_cmd
      curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --node-external-ip="${6}" --token="${8}" --datastore-endpoint="${9} ${10}"
   else
      echo "curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --node-external-ip=\"${6}\" --token \"${8}\" --datastore-endpoint=\"${9}\"" >/tmp/master_cmd
      curl -sfL https://get.bhojpur.net/dcp/install.sh | INSTALL_DCP_TYPE='server' sh -s - --node-external-ip="${6}" --token="${8}" --datastore-endpoint="${9}"
    fi
fi