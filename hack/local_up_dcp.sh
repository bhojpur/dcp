#!/usr/bin/env bash

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
#
# This shell will create a Bhojpur DCP cluster locally with kind. The dcp-tunnel will be
# automatically deployed, and the autonomous mode will be active.
#
# It uses the following env variables:
# REGION
# REGION affects the GOPROXY to use. You can set it to "in" to use GOPROXY="https://goproxy.in".
# Default value is "us", which means using GOPROXY="https://goproxy.io".
#
# KIND_KUBECONFIG
# KIND_KUBECONFIG represents the path to store the kubeconfig file of the cluster
# which is created by this shell. The default value is "$HOME/.kube/config".
#
# NODES_NUM
# NODES_NUM represents the number of nodes to set up in the new-created cluster.
# There is one control-plane node and NODES_NUM-1 worker nodes. Thus, NODES_NUM must
# not be less than 2. The default value is 2.
#
# KUBERNETESVERSION
# KUBERNETESVERSION declares the kubernetes version the cluster will use. The format is "v1.XX". 
# Now only v1.17, v1.18, v1.19, v1.20 and v1.21 are supported. The default value is v1.21.

set -x
set -e
set -u

DCP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"

readonly REQUIRED_CMD=(
    go
    docker
    kubectl
    kind
)

readonly BUILD_TARGETS=(
    dcpsvr
    controller-manager
    dcpctl
    tunnel-server
    tunnel-agent
    node-servant
)

readonly LOCAL_ARCH=$(go env GOHOSTARCH)
readonly LOCAL_OS=$(go env GOHOSTOS)
readonly CLUSTER_NAME="dcp-e2e-test"
readonly KUBERNETESVERSION=${KUBERNETESVERSION:-"v1.21"}
readonly NODES_NUM=${NODES_NUM:-2}
readonly KIND_KUBECONFIG=${KIND_KUBECONFIG:-${HOME}/.kube/config}

function install_kind {
    echo "Begin to install kind"
    GO111MODULE="on" go get sigs.k8s.io/kind@v0.11.1
}

function install_docker {
    echo "docker should be installed first"
    return -1
}

function install_kubectl {
    echo "kubectl should be installed first"
    return -1
} 

function install_go {
    echo "Go should be installed first"
    return -1
}

function preflight {
    echo "Preflight Check..."
    for bin in "${REQUIRED_CMD[@]}"; do
        command -v ${bin} > /dev/null 2>&1
        if [[ $? -ne 0 ]]; then
            echo "Cannot find command ${bin}."
            install_${bin}
            if [[ $? -ne 0 ]]; then
                echo "Error occurred, exit"
                exit -1
            fi
        fi
    done
}

function build_target_binaries_and_images {
    echo "Begin to build Bhojpur DCP binaries and images"

    export WHAT=${BUILD_TARGETS[@]}
    export ARCH=${LOCAL_ARCH}

    source ${DCP_ROOT}/hack/make-rules/release-images.sh    
}

function local_up_dcp {
    echo "Begin to setup Bhojpur DCP cluster"
    dcp_version=$(get_version ${LOCAL_ARCH})
    ${DCP_LOCAL_BIN_DIR}/${LOCAL_OS}/${LOCAL_ARCH}/dcpctl test init \
      --kubernetes-version=${KUBERNETESVERSION} --kube-config=${KIND_KUBECONFIG} \
      --cluster-name=${CLUSTER_NAME} --dcp-version=${dcp_version} --use-local-images --ignore-error \
      --node-num=${NODES_NUM}
}

function cleanup {
    rm -rf ${DCP_ROOT}/_output
    rm -rf ${DCP_ROOT}/dockerbuild
    kind delete clusters ${CLUSTER_NAME}
}

function cleanup_on_err {
    if [[ $? -ne 0 ]]; then
        cleanup
    fi
}


trap cleanup_on_err EXIT

cleanup
preflight
build_target_binaries_and_images
local_up_dcp