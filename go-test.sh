#!/bin/bash -euo pipefail

export GREP_OPTIONS='--color=always'
export AUDIO_TESTING=1
go clean -testcache
go test -v ./src/audio \
  | GREP_COLOR='1;31' grep -E '.*FAIL.*|$' \
  | GREP_COLOR='1;32' grep -E '.*PASS.*|$'
