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

set -o errexit
set -o nounset
set -o pipefail

DCP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
DCP_MOD="$(head -1 $DCP_ROOT/go.mod | awk '{print $2}')"
DCP_OUTPUT_DIR=${DCP_ROOT}/_output
DCP_BIN_DIR=${DCP_OUTPUT_DIR}/bin
DCP_LOCAL_BIN_DIR=${DCP_OUTPUT_DIR}/local/bin

PROJECT_PREFIX=${PROJECT_PREFIX:-dcp}
LABEL_PREFIX=${LABEL_PREFIX:-bhojpur.net}
GIT_VERSION=${GIT_VERSION:-$(git describe --abbrev=0 --tags)}
GIT_COMMIT=$(git rev-parse HEAD)
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
REPO=${REPO:-dcp}
TAG=$GIT_VERSION

source "${DCP_ROOT}/hack/lib/common.sh"
source "${DCP_ROOT}/hack/lib/build.sh"
source "${DCP_ROOT}/hack/lib/release-images.sh"
source "${DCP_ROOT}/hack/lib/release-manifest.sh"