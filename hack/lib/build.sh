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

readonly DCP_ALL_TARGETS=(
    dcpctl
    node-servant
    dcpsvr
    controller-manager
    tunnel-server
    tunnel-agent
)

# we will generates setup yaml files for following components
readonly DCP_YAML_TARGETS=(
    dcpsvr
    controller-manager
    tunnel-server
    tunnel-agent
)

#PROJECT_PREFIX=${PROJECT_PREFIX:-dcp}
#LABEL_PREFIX=${LABEL_PREFIX:-bhojpur.net}
#GIT_VERSION="v0.1.1"
#GIT_COMMIT=$(git rev-parse HEAD)
#BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

# project_info generates the project information and the corresponding value
# for 'ldflags -X' option
project_info() {
    PROJECT_INFO_PKG=${DCP_MOD}/pkg/projectinfo
    echo "-X ${PROJECT_INFO_PKG}.projectPrefix=${PROJECT_PREFIX}"
    echo "-X ${PROJECT_INFO_PKG}.labelPrefix=${LABEL_PREFIX}"
    echo "-X ${PROJECT_INFO_PKG}.gitVersion=${GIT_VERSION}"
    echo "-X ${PROJECT_INFO_PKG}.gitCommit=${GIT_COMMIT}"
    echo "-X ${PROJECT_INFO_PKG}.buildDate=${BUILD_DATE}"
}

# get_binary_dir_with_arch generated the binary's directory with GOOS and GOARCH.
# eg: ./_output/bin/darwin/arm64/
get_binary_dir_with_arch(){
    echo $1/$(go env GOOS)/$(go env GOARCH)/
}

build_binaries() {
    local goflags goldflags gcflags
    goldflags="${GOLDFLAGS:--s -w $(project_info)}"
    gcflags="${GOGCFLAGS:-}"
    goflags=${GOFLAGS:-}

    local -a targets=()
    local arg

    for arg; do
      if [[ "${arg}" == -* ]]; then
        # Assume arguments starting with a dash are flags to pass to go.
        goflags+=("${arg}")
      else
        targets+=("${arg}")
      fi
    done

    if [[ ${#targets[@]} -eq 0 ]]; then
      targets=("${DCP_ALL_TARGETS[@]}")
    fi

    local target_bin_dir=$(get_binary_dir_with_arch ${DCP_LOCAL_BIN_DIR})
    mkdir -p ${target_bin_dir}
    cd ${target_bin_dir}
    for binary in "${targets[@]}"; do
      echo "Building ${binary}"
      go build -o $(get_output_name $binary) \
          -ldflags "${goldflags:-}" \
          -gcflags "${gcflags:-}" ${goflags} $DCP_ROOT/cmd/grid/$(canonicalize_target $binary)
    done

    if [[ $(host_platform) == ${HOST_PLATFORM} ]]; then
      rm -f "${DCP_BIN_DIR}"
      ln -s "${target_bin_dir}" "${DCP_BIN_DIR}"
    fi
}

# gen_yamls generates yaml files for user specified components by
# subsituting the place holders with envs
gen_yamls() {
    local -a yaml_targets=()
    for arg; do
        # ignoring go flags
        [[ "$arg" == -* ]] && continue
        target=$(basename $arg)
        # only add target that is in the ${DCP_YAML_TARGETS} list
        if [[ "${DCP_YAML_TARGETS[@]}" =~ "$target" ]]; then
            yaml_targets+=("$target")
        fi
    done
    # if not specified, generate yaml for default yaml targets
    if [ ${#yaml_targets[@]} -eq 0 ]; then
        yaml_targets=("${DCP_YAML_TARGETS[@]}")
    fi
    echo $yaml_targets

    local yaml_dir=$DCP_OUTPUT_DIR/setup/
    mkdir -p $yaml_dir
    for yaml_target in "${yaml_targets[@]}"; do
        oup_file=${yaml_target/dcp/$PROJECT_PREFIX}
        echo "generating yaml file for $oup_file"
        sed "s|__project_prefix__|${PROJECT_PREFIX}|g;
        s|__label_prefix__|$LABEL_PREFIX|g;
        s|__repo__|$REPO|g;
        s|__tag__|$TAG|g;" \
            $DCP_ROOT/config/yaml-template/$yaml_target.yaml > \
            $yaml_dir/$oup_file.yaml
    done
}