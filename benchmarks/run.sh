mkdir results
DURATION=30
for f in *.txt
do
	echo "Processing $f"
    vegeta attack --targets $f -duration=${DURATION}s --output results/$f.bin &
    curl -o results/$f.heap.prof "http://localhost:6060/debug/pprof/heap?seconds=${DURATION}" &
    curl -o results/$f.cpu.prof "http://localhost:6060/debug/pprof/profile?seconds=${DURATION}" &
    curl -o results/$f.goroutine.prof "http://localhost:6060/debug/pprof/goroutine?seconds=${DURATION}" &
    curl -o results/$f.block.prof "http://localhost:6060/debug/pprof/block?seconds=${DURATION}" &
    curl -o results/$f.mutex.prof "http://localhost:6060/debug/pprof/mutex?seconds=${DURATION}" &
    # curl -o results/$f.trace.prof "http://localhost:6060/debug/pprof/trace" & # seconds doesn't work
    sleep ${DURATION} # wait for everything to finish
    sleep ${DURATION} # make double sure that everything is finished before moving on
done