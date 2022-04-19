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

# get_output_name generates the executable's name. If the $PROJECT_PREFIX
# is set, it subsitutes the prefix of the executable's name with the env,
# otherwise the basename of the target is used
get_output_name() {
    local oup_name=$(canonicalize_target $1)
    PROJECT_PREFIX=${PROJECT_PREFIX:-}
    if [ -z $PROJECT_PREFIX ]; then
        oup_name=${oup_name}
    elif [ "$PROJECT_PREFIX" = "dcp" ]; then
        oup_name=${oup_name}
    else
        oup_name=${oup_name/dcp-/$PROJECT_PREFIX}
        oup_name=${oup_name/dcp/$PROJECT_PREFIX}
    fi
    echo $oup_name
}

# canonicalize_target delete the first four characters when
# target begins with "cmd/"
canonicalize_target() {
    local target=$1
    if [[ "$target" =~ ^cmd/.* ]]; then
        target=${target:4}
    fi

    echo $target
}

# host_platform returns the host platform determined by golang
host_platform() {
  echo "$(go env GOHOSTOS)/$(go env GOHOSTARCH)"
}

# Parameters
# $1: binary_name
get_component_name() {
  local dcp_component_name
  if [[ $1 =~ dcpctl ]]
  then
    dcp_component_name="dcpctl-servant"
  elif [[ $1 =~ dcp-node-servant ]];
  then
    dcp_component_name="node-servant"
  else
    dcp_component_name=$1
  fi
  echo $dcp_component_name
}