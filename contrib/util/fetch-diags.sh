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

# echo -e 'bhojpur-dev-file.json\ndcp-diagnostic-logs' | gsutil config -e
if ! command -v gsutil >/dev/null 2>&1; then
  echo "gsutil command not found" >&2
  echo "hint: pip install gsutil" >&2
  exit 1
fi

UUID=${1:-}
if [ -z "$UUID" ]; then
  echo "ERROR: should pass UUID as first arg"
  exit 1
fi

PRIVATE_PEM=${PRIVATE_PEM:-'diags.private.pem'}
PRIVATE_KEY=${PRIVATE_KEY:-'diags.private.key'}

if [ ! -f "$PRIVATE_PEM" ]; then
  echo "WARNING: PRIVATE_PEM $PRIVATE_PEM not found" >&2
fi
if [ ! -f "$PRIVATE_KEY" ]; then
  echo "WARNING: PRIVATE_KEY $PRIVATE_KEY not found (PRIVATE_PEM password)" >&2
fi

DIAGPROG=${DIAGPROG:-'dcp'}
BUCKET_NAME=${BUCKET_NAME:-"$DIAGPROG-diagnostic-logs"}

decrypt() {
  if [ -z "$1" ]; then
    echo "ERROR: decrypt param undefined" >&2
    return 1
  fi

  if [ ! -f "$1.tar.gz" ]; then
    echo "ERROR: $1.tar.gz does not exist" >&2
    return 1
  fi

  tar xzf "$1.tar.gz"
  if [ ! -f "$1.logs.tar.gz.enc" ]; then
    echo "ERROR: encrypted logs file $1.logs.tar.gz.enc does not exist" >&2
    return 1
  fi

  if [ -f "$1.meta.enc" ]; then
    openssl rsautl -decrypt -passin "file:$PRIVATE_KEY" -inkey "$PRIVATE_PEM" -in "$1.meta.enc" -out "$1.meta"
  fi

  local salt="$salt"
  local key="$key"
  local iv="$iv"

  if [ -f "$1.meta" ]; then
    salt=$(grep 'salt=' $1.meta | cut -f2 -d=)
    key=$(grep 'key=' $1.meta | cut -f2 -d=)
    iv=$(grep 'iv=' $1.meta | cut -f2 -d=)
  fi

  if [ -z "$salt" ] || [ -z "$key" ] || [ -z "$iv" ]; then
    echo "$1: Missing decryption metadata" >&2
    echo
    return 1
  fi

  if openssl enc -d -aes-256-cbc -in "$1.logs.tar.gz.enc" -out /dev/stdout -S "$salt" -K "$key" -iv "$iv" | tar xzf -; then
    rm "$1".*
    return 0
  fi

  echo "$1: Error decrypting" >&2
  echo
  return 1
}

fetch() {
  for url in $@; do
    local log=$(basename "$url" | cut -f1 -d.)
    echo "Downloading $log"
    gsutil cp $url $log.tar.gz
    decrypt $log
  done
}

{
  fetch $(gsutil ls "gs://$BUCKET_NAME/*$UUID*")
}