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

# Grabs the last 5 commit SHA's from the given branch, then purges any commits that do not have a passing CI build
iterations=0
curl -s -H 'Accept: application/vnd.github.v3+json' "https://api.github.com/repos/bhojpur/dcp/commits?per_page=5&sha=$1" | jq -r '.[] | .sha'  &> $2
# The VMs take time on startup to hit googleapis.com, wait loop until we can
while ! curl -s --fail https://storage.googleapis.com/bhojpur-net-platform > /dev/null; do
    ((iterations++))
    if [ "$iterations" -ge 30 ]; then
        echo "Unable to hit googleapis.com/bhojpur-net-platform"
        exit 1
    fi
    sleep 1
done

iterations=0
curl -s --fail https://storage.googleapis.com/bhojpur-net-platform/dcp-$(head -n 1 $2).sha256sum
while [ $? -ne 0 ]; do
    ((iterations++))
    if [ "$iterations" -ge 6 ]; then
        echo "No valid commits found"
        exit 1
    fi
    sed -i 1d "$2"
    sleep 1
    curl -s --fail https://storage.googleapis.com/bhojpur-net-platform/dcp-$(head -n 1 $2).sha256sum
done