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

set -x

DCP_IMAGE_DIR=${DCP_OUTPUT_DIR}/images
DCPCTL_SERVANT_DIR=${DCP_ROOT}/config/dcpctl-servant
DOCKER_BUILD_BASE_DIR=$DCP_ROOT/dockerbuild
DCP_BUILD_IMAGE="golang:1.17.5-alpine3.15"
#REPO="dcp"
#TAG="v0.2.0"

readonly -a DCP_BIN_TARGETS=(
    dcpsvr
    controller-manager
    dcpctl
    node-servant
    tunnel-server
    tunnel-agent
)

readonly -a SUPPORTED_ARCH=(
    amd64
    arm
    arm64
)

readonly SUPPORTED_OS=linux

readonly -a bin_targets=(${WHAT[@]:-${DCP_BIN_TARGETS[@]}})
readonly -a bin_targets_process_servant=("${bin_targets[@]/dcpctl-servant/dcpctl}")
readonly -a target_arch=(${ARCH[@]:-${SUPPORTED_ARCH[@]}})
readonly region=${REGION:-us}

# Parameters
# $1: component name
# $2: arch
function get_image_name {
    tag=$(get_version $2)
    echo "${REPO}/$1:${tag}"
}

# Parameters
# $1: arch
# The format is like: 
# "v0.6.0-amd64-a955ecc" if the HEAD is not at a tag,
# "v0.6.0-amd64" otherwise.
function get_version {
    # If ${GIT_COMMIT} does not point at a tag, add commit suffix to the image tag.
    if [[ -z $(git tag --points-at ${GIT_COMMIT}) ]]; then
        tag="${TAG}-$1-$(echo ${GIT_COMMIT} | cut -c 1-7)"
    else
        tag="${TAG}-$1"
    fi    

    echo "${tag}"
}


function build_multi_arch_binaries() {
    local docker_dcp_root="/opt/src"
    local docker_run_opts=(
        "-i"
        "--rm"
        "--network host"
        "-v ${DCP_ROOT}:${docker_dcp_root}"
        "--env CGO_ENABLED=0"
        "--env GOOS=${SUPPORTED_OS}"
        "--env PROJECT_PREFIX=${PROJECT_PREFIX}"
        "--env LABEL_PREFIX=${LABEL_PREFIX}"
        "--env GIT_VERSION=${GIT_VERSION}"
        "--env GIT_COMMIT=${GIT_COMMIT}"
        "--env BUILD_DATE=${BUILD_DATE}"
        "--env HOST_PLATFORM=$(host_platform)"
    )
    # use goproxy if build from inside India
    [[ $region == "in" ]] && docker_run_opts+=("--env GOPROXY=https://goproxy.in")

    # use proxy if set
    [[ -n ${http_proxy+x} ]] && docker_run_opts+=("--env http_proxy=${http_proxy}")
    [[ -n ${https_proxy+x} ]] && docker_run_opts+=("--env https_proxy=${https_proxy}")

    local docker_run_cmd=(
        "/bin/sh"
        "-xe"
        "-c"
    )

    local sub_commands="sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories; \
        apk --no-cache add bash git; \
        cd ${docker_dcp_root}; umask 0022; \
        rm -rf ${DCP_LOCAL_BIN_DIR}/* ; \
        git config --global --add safe.directory ${docker_dcp_root};"
    for arch in ${target_arch[@]}; do
        sub_commands+="GOARCH=$arch bash ./hack/make-rules/build.sh $(echo ${bin_targets_process_servant[@]}); "
    done
    sub_commands+="chown -R $(id -u):$(id -g) ${docker_dcp_root}/_output"

    docker run ${docker_run_opts[@]} ${DCP_BUILD_IMAGE} ${docker_run_cmd[@]} "${sub_commands}"
}

function build_docker_image() {
    for arch in ${target_arch[@]}; do
        for binary in "${bin_targets_process_servant[@]}"; do
           local binary_name=$(get_output_name $binary)
           local binary_path=${DCP_LOCAL_BIN_DIR}/${SUPPORTED_OS}/${arch}/${binary_name}
           if [ -f ${binary_path} ]; then
               local docker_build_path=${DOCKER_BUILD_BASE_DIR}/${SUPPORTED_OS}/${arch}
               local docker_file_path=${docker_build_path}/Dockerfile.${binary_name}-${arch}
               mkdir -p ${docker_build_path}
               local dcp_component_name=$(get_component_name $binary_name)
               local base_image
               if [[ ${binary} =~ dcpctl ]]
               then
                 case $arch in
                  amd64)
                      base_image="amd64/alpine:3.9"
                      ;;
                  arm64)
                      base_image="arm64v8/alpine:3.9"
                      ;;
                  arm)
                      base_image="arm32v7/alpine:3.9"
                      ;;
                  *)
                      echo unknown arch $arch
                      exit 1
                 esac
                 cat << EOF > $docker_file_path
FROM ${base_image}
ADD ${binary_name} /usr/local/bin/dcpctl
EOF
               elif [[ ${binary} =~ node-servant ]];
               then
                 case $arch in
                  amd64)
                      base_image="amd64/alpine:3.9"
                      ;;
                  arm64)
                      base_image="arm64v8/alpine:3.9"
                      ;;
                  arm)
                      base_image="arm32v7/alpine:3.9"
                      ;;
                  *)
                      echo unknown arch $arch
                      exit 1
                 esac
                 ln ./hack/lib/node-servant-entry.sh "${docker_build_path}/entry.sh"
                 cat << EOF > $docker_file_path
FROM ${base_image}
ADD entry.sh /usr/local/bin/entry.sh
RUN chmod +x /usr/local/bin/entry.sh
ADD ${binary_name} /usr/local/bin/node-servant
EOF
               else
                 base_image="k8s.gcr.io/debian-iptables-${arch}:v11.0.2"
                 cat <<EOF > "${docker_file_path}"
FROM ${base_image}
COPY ${binary_name} /usr/local/bin/${binary_name}
ENTRYPOINT ["/usr/local/bin/${binary_name}"]
EOF
               fi

               dcp_component_image=$(get_image_name ${dcp_component_name} ${arch})
               ln "${binary_path}" "${docker_build_path}/${binary_name}"
               docker build --no-cache -t "${dcp_component_image}" -f "${docker_file_path}" ${docker_build_path}
               echo ${dcp_component_image} >> ${DOCKER_BUILD_BASE_DIR}/images.list
               docker save ${dcp_component_image} > ${DCP_IMAGE_DIR}/${dcp_component_name}-${SUPPORTED_OS}-${arch}.tar
            fi
        done
    done
}

build_images() {
    # Always clean first
    rm -Rf ${DCP_OUTPUT_DIR}
    rm -Rf ${DOCKER_BUILD_BASE_DIR}
    mkdir -p ${DCP_LOCAL_BIN_DIR}
    mkdir -p ${DCP_IMAGE_DIR}
    mkdir -p ${DOCKER_BUILD_BASE_DIR}

    build_multi_arch_binaries
    build_docker_image
}

push_images() {
    cat ${DOCKER_BUILD_BASE_DIR}/images.list | xargs -I % sh -c 'echo pushing %; docker push %; echo'
}