#!/bin/bash

if [ $(id -u) = 0 ]; then
    BHOJPUR_PATH="/var/lib/bhojpur"
else
    BHOJPUR_PATH="$HOME/.bhojpur"
fi