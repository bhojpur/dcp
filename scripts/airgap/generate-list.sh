#!/bin/bash
set -e -x

cd $(dirname $0)

dcp crictl images -o json \
    | jq -r '.images[].repoTags[0] | select(. != null)' \
    | tee image-list.txt