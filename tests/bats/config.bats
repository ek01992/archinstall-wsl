#!/usr/bin/env bats

load 'test_helper'

setup() {
  REPO_ROOT="$BATS_TEST_DIRNAME/../.."
  ROOT_DIR="$REPO_ROOT"
  source "$REPO_ROOT/lib/bash/config.sh"
}

@test "split_words_to_array splits space-delimited string" {
  run bash -c 'mapfile -t arr < <(split_words_to_array "a b  c"); echo ${#arr[@]}; echo ${arr[0]}:${arr[1]}:${arr[2]}'
  [ "$status" -eq 0 ]
  # length
  [ "$(echo "$output" | sed -n '1p')" = "3" ]
  # elements
  [ "$(echo "$output" | sed -n '2p')" = "a:b:c" ]
}
