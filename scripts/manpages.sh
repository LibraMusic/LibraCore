#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
go run ./cmd/libra man | gzip -c -9 >manpages/libra.1.gz
