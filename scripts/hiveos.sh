#!/usr/bin/env bash

# automatically set the version in hiveoses h-manifest.conf file.
# The version is read from the first argument passed to the script.

if [ -z "$1" ]; then
    echo 'Missing version' >&2
    exit 1
fi

sed -i "/MINER_LATEST_VER/c MINER_LATEST_VER=$1" hiveos/h-manifest.conf