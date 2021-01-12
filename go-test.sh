#!/bin/bash -euo pipefail

export GREP_OPTIONS='--color=always'

go test "$@" \
  | GREP_COLOR='1;31' grep -E '.*FAIL.*|$' \
  | GREP_COLOR='1;32' grep -E '.*PASS.*|$'
