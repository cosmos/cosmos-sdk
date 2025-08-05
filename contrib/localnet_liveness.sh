#!/bin/bash

CNT=0
ITER=$1
SLEEP=$2
NUMBLOCKS=$3
NODEADDR=$4

if [ -z "$1" ]; then
  echo "Need to input number of iterations to run..."
  exit 1
fi

if [ -z "$2" ]; then
  echo "Need to input number of seconds to sleep between iterations"
  exit 1
fi

if [ -z "$3" ]; then
  echo "Need to input block height to declare completion..."
  exit 1
fi

if [ -z "$4" ]; then
  echo "Need to input node address to poll..."
  exit 1
fi

docker_containers=( $(docker ps -f name=simd --format='{{.Names}}') )
echo "Found ${#docker_containers[@]} docker containers matching 'simd'"

while [ ${CNT} -lt $ITER ]; do
  echo "Iteration $((CNT+1))/$ITER - Polling $NODEADDR:26657/status"
  response=$(curl -s $NODEADDR:26657/status)
  curl_exit_code=$?

  if [ $curl_exit_code -ne 0 ]; then
    echo "Error: curl failed with exit code $curl_exit_code"
  elif [ -z "$response" ]; then
    echo "Error: Empty response from $NODEADDR:26657/status"
  else
    curr_block=$(echo "$response" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null)
    if [ $? -ne 0 ]; then
      echo "Error: Failed to parse JSON response"
      echo "Response: $response"
    elif [ ! -z ${curr_block} ] && [ "${curr_block}" != "null" ]; then
      echo "Number of Blocks: ${curr_block}"
    else
      echo "Warning: No block height found in response"
    fi
  fi

  if [ ! -z ${curr_block} ] && [ ${curr_block} -gt ${NUMBLOCKS} ]; then
    echo "Number of blocks reached. Success!"
    exit 0
  fi

  # Emulate network chaos:
  #
  # Every 10 blocks, pick a random container and restart it.
  if ! ((${CNT} % 10)) && [ ${#docker_containers[@]} -gt 0 ]; then
    rand_container=${docker_containers["$[RANDOM % ${#docker_containers[@]}]"]};
    echo "Restarting random docker container ${rand_container}"
    docker restart ${rand_container} &>/dev/null &
  fi
  let CNT=CNT+1
  sleep $SLEEP
done
echo "Timeout reached. Failure!"
exit 1
