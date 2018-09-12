#!/bin/bash

tmpdir=$(mktemp -d)
echo "Using temporary log directory: $tmpdir"

sim() {
  seed=$1
	echo "Running full Gaia simulation with seed $seed. This may take awhile!"
  file="$tmpdir/gaia-simulation-seed-$seed-date-$(date -Iseconds -u).stdout"
  echo "Writing stdout to $file..."
	go test ./cmd/gaia/app -run TestFullGaiaSimulation -SimulationEnabled=true -SimulationNumBlocks=1000 -SimulationVerbose=true -SimulationSeed=$seed -v -timeout 24h > $file
}

seeds=(1 2 4)
# seeds=(1 2 4 7 9 20 32 123 4728 37827 981928 87821 891823782 989182 89182391)
i=0
pids=()
for seed in ${seeds[@]}; do
  sim $seed &
  pids[${i}]=$!
  i=$(($i+1))
  sleep 0.1 # start in order, nicer logs
done

echo "Simulation processes spawned, waiting for completion..."

code=0

for pid in ${pids[*]}; do
  wait $pid
  last=$?
  if [ $last -ne 0 ]; then
    seed=$seeds[${i}]
    echo "Simulation with seed $seed failed!"
    code=1
  fi
done

exit $code
