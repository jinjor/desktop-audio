#!/bin/bash
set -euo pipefail

export AUDIO_TESTING=1

grep='grep --color=always'
go clean -testcache
go test -v ./src/audio \
  | GREP_COLOR='1;31' grep -E '.*FAIL.*|$' \
  | GREP_COLOR='1;32' grep -E '.*PASS.*|$'
