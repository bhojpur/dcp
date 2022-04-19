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

# Parameters
# $1: component name
function get_manifest_name() {
      # If ${GIT_COMMIT} is not at a tag, add commit to the image tag.
      if [[ -z $(git tag --points-at ${GIT_COMMIT}) ]]; then
          dcp_component_manifest="${REPO}/$1:${TAG}-$(echo ${GIT_COMMIT} | cut -c 1-7)"
      else
          dcp_component_manifest="${REPO}/$1:${TAG}"
      fi
      echo ${dcp_component_manifest}
}

function build_docker_manifest() {
    # Always clean first
    rm -Rf ${DOCKER_BUILD_BASE_DIR}
    mkdir -p ${DOCKER_BUILD_BASE_DIR}

    for binary in "${bin_targets_process_servant[@]}"; do
      local binary_name=$(get_output_name $binary)
      local dcp_component_name=$(get_component_name $binary_name)
      local dcp_component_manifest=$(get_manifest_name $dcp_component_name)
      echo ${dcp_component_manifest} >> ${DOCKER_BUILD_BASE_DIR}/manifest.list
      # Remove existing manifest.
      docker manifest rm ${dcp_component_manifest} || true
      for arch in ${target_arch[@]}; do
        case $arch in
          amd64)
              ;;
          arm64)
              ;;
          arm)
              ;;
          *)
              echo unknown arch $arch
              exit 1
         esac
         dcp_component_image=$(get_image_name ${dcp_component_name} ${arch})
         docker manifest create ${dcp_component_manifest} --amend  ${dcp_component_image}
         docker manifest annotate ${dcp_component_manifest} ${dcp_component_image} --os ${SUPPORTED_OS} --arch ${arch}
      done
    done
}

push_manifest() {
    cat ${DOCKER_BUILD_BASE_DIR}/manifest.list | xargs -I % sh -c 'echo pushing manifest %; docker manifest push --purge %; echo'
}