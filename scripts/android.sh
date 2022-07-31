#!/bin/sh
set -e

BASE="https://github.com/whalesburg/derocli"
LATEST=$(curl -fsSLI -o /dev/null -w %{url_effective} $BASE/releases/latest)
VERSION=$(basename $LATEST)
ARCHIVE="derocli-$VERSION-linux-armv7.tar.gz"

curl -sLJO "$BASE/releases/download/$VERSION/$ARCHIVE"


