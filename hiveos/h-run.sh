#!/usr/bin/env bash

[[ ! -e ./dero-stratum-miner.conf ]] && echo "No config file found, exiting" && pwd && exit 1

dero-stratum-miner $(< dero-stratum-miner.conf) | tee --append $CUSTOM_LOG_BASENAME.log
