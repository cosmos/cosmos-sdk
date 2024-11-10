#!/bin/bash

# Number of processes to start
n=$1

# Function to start a load-test process
start_process() {
  simdv2 tx benchmark load-test --from $1 --yes &
  echo $!
}

# Trap SIGINT to kill all child processes
trap 'kill $(jobs -p)' SIGINT

files=(~/.simappv2/keyring-test/*.info)
if [ ${#files[@]} -lt $n ]; then
  echo "Error: Not enough accounts. Found ${#files[@]}, but need $n."
  exit 1
fi

for i in $(seq 0 $(($n - 1))); do
  echo "Selected account: ${files[$i]}"
  start_process `basename ${files[$i]} .info`
done

# Wait for all processes to complete
wait
