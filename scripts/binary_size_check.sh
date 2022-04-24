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

set -e

if [ "${DEBUG}" = 1 ]; then
    set -x
fi

. ./scripts/version.sh

# Try to keep the Bhojpur DCP binary image under 64 megabytes.
# "64M ought to be enough for anybody"
MAX_BINARY_MB=64
MAX_BINARY_SIZE=$((MAX_BINARY_MB * 1024 * 1024))
BIN_SUFFIX="-${ARCH}"
if [ ${ARCH} = amd64 ]; then
    BIN_SUFFIX=""
elif [ ${ARCH} = arm ]; then
    BIN_SUFFIX="-armhf"
fi

CMD_NAME="dist/artifacts/dcp${BIN_SUFFIX}"
SIZE=$(stat -c '%s' ${CMD_NAME})

if [ ${SIZE} -gt ${MAX_BINARY_SIZE} ]; then
  echo "Bhojpur DCP binary ${CMD_NAME} size ${SIZE} exceeds max acceptable size of ${MAX_BINARY_SIZE} bytes (${MAX_BINARY_MB} MiB)"
  exit 1
fi

echo "Bhojpur DCP binary ${CMD_NAME} size ${SIZE} is less than max acceptable size of ${MAX_BINARY_SIZE} bytes (${MAX_BINARY_MB} MiB)"
exit 0