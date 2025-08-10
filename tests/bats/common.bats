#!/usr/bin/env bats

load 'test_helper'

setup() {
  REPO_ROOT="$BATS_TEST_DIRNAME/../.."
  ROOT_DIR="$REPO_ROOT"
  source "$REPO_ROOT/lib/bash/common.sh"
  source "$REPO_ROOT/lib/bash/config.sh"
}

@test "append_once writes line only once" {
  tmpfile="$BATS_TMPDIR/ao.txt"
  run bash -c 'append_once "hello" "$tmpfile"; append_once "hello" "$tmpfile"; cat "$tmpfile"'
  [ "$status" -eq 0 ]
  [ "$output" = "hello" ]
}

@test "normalize_bool maps values correctly" {
  run bash -c 'normalize_bool yes'
  [ "$output" = "true" ]
  run bash -c 'normalize_bool NO'
  [ "$output" = "false" ]
  run bash -c 'normalize_bool ""'
  [ "$output" = "false" ]
}
