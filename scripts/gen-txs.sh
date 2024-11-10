#!/bin/bash

# Number of processes to start
n=$1

# Function to start a load-test process
start_process() {
  simdv2 tx benchmark load-test alice --broadcast-mode async --yes &
  echo $!
}

# Trap SIGINT to kill all child processes
trap 'kill $(jobs -p)' SIGINT

# Start n processes
for i in $(seq 1 $n); do
  start_process
  sleep 1
done

# Wait for all processes to complete
wait
