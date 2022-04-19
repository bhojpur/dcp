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
set -e

DCP_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

TMP_DIR=$(mktemp -d)
mkdir -p "${TMP_DIR}"/src/github.com/bhojpur/dcp/pkg/appmanager/client
cp -r ${DCP_ROOT}/{go.mod,go.sum} "${TMP_DIR}"/src/github.com/bhojpur/dcp/
cp -r ${DCP_ROOT}/pkg/appmanager/{apis,hack} "${TMP_DIR}"/src/github.com/bhojpur/dcp/pkg/appmanager/

(
  cd "${TMP_DIR}"/src/github.com/bhojpur/dcp/;
  HOLD_GO="${TMP_DIR}/src/github.com/bhojpur/dcp/pkg/appmanager/hack/hold.go"
  printf 'package hack\nimport "k8s.io/code-generator"\n' > ${HOLD_GO}
  go mod vendor
  GOPATH=${TMP_DIR} GO111MODULE=off /bin/bash vendor/k8s.io/code-generator/generate-groups.sh all \
    github.com/bhojpur/dcp/pkg/appmanager/client github.com/bhojpur/dcp/pkg/appmanager/apis apps:v1alpha1 -h ./pkg/appmanager/hack/boilerplate.go.txt
)

rm -rf ./pkg/appmanager/client/{clientset,informers,listers}
mv "${TMP_DIR}"/src/github.com/bhojpur/dcp/pkg/appmanager/client/* ./pkg/appmanager/client
