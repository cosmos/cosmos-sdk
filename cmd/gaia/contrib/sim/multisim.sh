#!/bin/bash

seeds=(1 2 4 7 9 20 32 123 124 582 1893 2989 3012 4728 37827 981928 87821 891823782 989182 89182391 \
11 22 44 77 99 2020 3232 123123 124124 582582 18931893 29892989 30123012 47284728 37827)
blocks=$1
period=$2
testname=$3
genesis=$4

f_echo_stderr() {
  echo $@ >&2
}

cleanup() {
  local l_children

  l_children=$(ps -o pid= --ppid $$)
  echo "Stopping children ["${l_children}"] ..." >&2
  kill -SIGSTOP ${l_children} || true
  echo "Terminating children ["${l_children}"] ..." >&2
  kill -TERM ${l_children} || true
  exit 0
}

trap cleanup SIGINT SIGTERM

f_sim() {
  local l_cmd l_seed=$1

  f_echo_stderr "Running Gaia simulation with seed $seed. This may take awhile!"
  file="$tmpdir/gaia-simulation-seed-$seed-date-$(date -u +"%Y-%m-%dT%H:%M:%S+00:00").stdout"
  f_echo_stderr "Redirecting stdout to $file..."

  l_cmd="go test github.com/cosmos/cosmos-sdk/cmd/gaia/app -run $testname -SimulationEnabled=true -SimulationNumBlocks=$blocks -SimulationGenesis=$genesis \
    -SimulationVerbose=true -SimulationCommit=true -SimulationSeed=$seed -SimulationPeriod=$period -v -timeout 24h"
  ${l_cmd} > $file || echo -e "Simulation with seed $seed failed!\nTo replicate, run 'go test ./cmd/gaia/app -run $testname -SimulationEnabled=true -SimulationNumBlocks=$blocks -SimulationVerbose=true -SimulationCommit=true -SimulationSeed=$seed -v -timeout 24h'"
}

f_spinner() {
  local l_i l_sp
  l_i=1
  l_sp="/-\|"
  echo -n ' '
  while true
  do
    printf "\b${l_sp:l_i++%${#l_sp}:1}"
    sleep 1s
  done
}

go mod download >&2

f_echo_stderr "Running multi-seed simulation with seeds ${seeds[@]}"
f_echo_stderr "Running $blocks blocks per seed"
f_echo_stderr "Running test $testname"
f_echo_stderr "Using genesis file $genesis"
f_echo_stderr "Edit scripts/multisim.sh to add new seeds. Keeping parameters in the file makes failures easy to reproduce."
f_echo_stderr "This script will kill all sub-simulations on SIGINT/SIGTERM (i.e. Ctrl-C)."

tmpdir=$(mktemp -d)
f_echo_stderr "Using temporary log directory: $tmpdir"

commands_file=$(mktemp)
echo Commands file $commands_file
for seed in ${seeds[@]}; do
  file="$tmpdir/gaia-simulation-seed-$seed-date-$(date -u +"%Y-%m-%dT%H:%M:%S+00:00").stdout"
  cmd="go test github.com/cosmos/cosmos-sdk/cmd/gaia/app -run $testname -SimulationEnabled=true -SimulationNumBlocks=$blocks -SimulationGenesis=$genesis \
     -SimulationVerbose=true -SimulationCommit=true -SimulationSeed=$seed -SimulationPeriod=$period -v -timeout 24h"
  echo "echo Run: ${cmd} ; echo Output: ${file} ; ${cmd} >$file ; echo Exit: \$?" >> ${commands_file} # || echo -e "Simulation with seed $seed failed!\nTo replicate, run 'go test ./cmd/gaia/app -run $testname -SimulationEnabled=true -SimulationNumBlocks=$blocks -SimulationVerbose=true -SimulationCommit=true -SimulationSeed=$seed -v -timeout 24h'" $seed \
done

f_echo_stderr "Simulation processes spawned, waiting for completion..."

f_spinner &
while read -r line; do sem --will-cite -j+0 "$line" ; done < ${commands_file}
sem --wait --will-cite
