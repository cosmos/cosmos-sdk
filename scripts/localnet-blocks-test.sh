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

while [ ${CNT} -lt $ITER ]; do
  var=$(curl -s $NODEADDR:26657/status | jq -r '.result.sync_info.latest_block_height')
  echo "Number of Blocks: ${var}"
  if [ ! -z ${var} ] && [ ${var} -gt ${NUMBLOCKS} ]; then
    echo "Number of blocks reached, exiting success..."
    exit 0
  fi
  let CNT=CNT+1
  sleep $SLEEP
done

echo "Timeout reached, exiting failure..."
exit 1
