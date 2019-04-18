#!/bin/bash

set -e

go mod download

seeds=(1 2 4 7 9 20 32 123 124 582 1893 2989 3012 4728 37827 981928 87821 891823782 989182 89182391 \
11 22 44 77 99 2020 3232 123123 124124 582582 18931893 29892989 30123012 47284728 37827)
blocks=$1
period=$2
testname=$3
genesis=$4

echo "Running multi-seed simulation with seeds ${seeds[@]}"
echo "Running $blocks blocks per seed"
echo "Running test $testname"
echo "Using genesis file $genesis"
echo "Edit scripts/multisim.sh to add new seeds. Keeping parameters in the file makes failures easy to reproduce."
echo "This script will kill all sub-simulations on SIGINT/SIGTERM (i.e. Ctrl-C)."

# Symbols prefixes legenda:
# f_ - function
# l_ - local symbols

f_spinner() {
  local l_i l_sp
  l_i=1
  l_chars="/-\|"
  echo -n ' '
  while true
  do
    printf "\b${l_chars:l_i++%${#l_chars}:1}"
    sleep 1s
  done
}

f_cleanup() {
  local l_children
  l_children=$(pgrep -P $$)
  echo "Stopping children ["${l_children}"] ..." >&2
  kill -SIGSTOP ${l_children} || true
  echo "Terminating children ["${l_children}"] ..." >&2
  kill -TERM ${l_children} || true
  exit 0
}

trap f_cleanup SIGINT SIGTERM EXIT

tmpdir=$(mktemp -d)
echo "Using temporary log directory: $tmpdir"

f_sim() {
  local l_seed
  l_seed=$1
	echo "Running Gaia simulation with seed $l_seed. This may take awhile!"
  file="$tmpdir/gaia-simulation-seed-$l_seed-date-$(date -u +"%Y-%m-%dT%H:%M:%S+00:00").stdout"
  echo "Writing stdout to $file..."
	go test github.com/cosmos/cosmos-sdk/cmd/gaia/app -run $testname -SimulationEnabled=true -SimulationNumBlocks=$blocks -SimulationGenesis=$genesis \
    -SimulationVerbose=true -SimulationCommit=true -SimulationSeed=$l_seed -SimulationPeriod=$period -v -timeout 24h > $file && \
    echo "Simulation with seed $l_seed OK" || ( code=1 ; echo "Simulation with seed $seed failed!" ) # > $file
}

echo "Simulation processes spawned, waiting for completion..."

f_spinner &


i=0
pids=()
for seed in ${seeds[@]}; do
  f_sim $seed &
  pids[${i}]=$!
  i=$(($i+1))
  sleep 10 # start in order, nicer logs
done

code=0
i=0
for pid in ${pids[*]}; do
  wait $pid
  last=$?
  seed=${seeds[${i}]}
  if [ $last -ne 0 ]
  then
    echo "Simulation with seed $seed failed!"
    echo "To replicate, run 'go test ./cmd/gaia/app -run $testname -SimulationEnabled=true -SimulationNumBlocks=$blocks -SimulationVerbose=true -SimulationCommit=true -SimulationSeed=$seed -v -timeout 24h'"
    code=1
  else
    echo "Simulation with seed $seed OK"
  fi
  i=$(($i+1))
done

exit $code
