#!/bin/bash

# Number of processes to start
n=$1

# Function to start a load-test process
start_process() {
  simd tx benchmark load-test --from "$1" --yes --ops 50 &
  echo $!
}

# Trap SIGINT to kill all child processes
trap 'kill $(jobs -p)' SIGINT

files=(~/.simapp/keyring-test/*.info)
if [ ${#files[@]} -lt "$n" ]; then
  echo "Error: Not enough accounts. Found ${#files[@]}, but need $n."
  exit 1
fi

for i in $(seq 0 $(($n - 1))); do
  echo "Selected account: ${files[$i]}"
  start_process "$(basename "${files[$i]}" .info)"
done

# Wait for all processes to complete
wait
