#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
go run . man | gzip -c >manpages/dero-stratum-miner.1.gz