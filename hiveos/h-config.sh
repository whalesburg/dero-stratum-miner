#!/usr/bin/env bash

local MINER_VER=$CUSTOM_VER
[[ -z $MINER_VER ]] && MINER_VER=$MINER_LATEST_VER
echo $MINER_VER

local MINER_VER=`miner_ver`

MINER_CONFIG="/hive/miners/custom/dero-stratum-miner/dero-stratum-miner.conf"
mkfile_from_symlink $MINER_CONFIG

#ALGO

#POOL
conf+="-r $CUSTOM_URL "
conf+="-w $CUSTOM_TEMPLATE "


# TLS and other options
conf+="$CUSTOM_USER_CONFIG --api-enabled --api-transport http --api-listen 127.0.0.1:44000"

echo -e "$conf" > $MINER_CONFIG
