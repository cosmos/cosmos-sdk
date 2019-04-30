package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	// default seeds
	seeds = []int{
		1, 2, 4, 7, 32, 123, 124, 582, 1893, 2989,
		3012, 4728, 37827, 981928, 87821, 891823782,
		989182, 89182391, 11, 22, 44, 77, 99, 2020,
		3232, 123123, 124124, 582582, 18931893,
		29892989, 30123012, 47284728,
	}

	// goroutine-safe process map
	procs map[int]*os.Process
	mutex *sync.Mutex

	// results channel
	results chan bool

	// command line arguments and options
	jobs     int
	blocks   string
	period   string
	testname string
	genesis  string

	// logs temporary directory
	tempdir string
)

func init() {
	log.SetPrefix("")
	log.SetFlags(0)

	procs = map[int]*os.Process{}
	mutex = &sync.Mutex{}
	flag.IntVar(&jobs, "j", 10, "Number of parallel processes")
	flag.StringVar(&genesis, "g", "", "Genesis file")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [-j maxprocs] [-g genesis.json] [blocks] [period] [testname]
Run simulations in parallel

`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func main() {
	var err error

	flag.Parse()
	if flag.NArg() != 3 {
		log.Fatal("wrong number of arguments")
	}

	// prepare input channel
	queue := make(chan int, len(seeds))
	for _, seed := range seeds {
		queue <- seed
	}
	close(queue)

	// jobs cannot be > len(seeds)
	if jobs > len(seeds) {
		jobs = len(seeds)
	}
	results = make(chan bool, len(seeds))

	// setup signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		fmt.Println()

		// drain the queue
		log.Printf("Draining seeds queue...")
		for seed := range queue {
			log.Printf("%d", seed)
		}
		log.Printf("Kill all remaining processes...")
		killAllProcs()
		os.Exit(1)
	}()

	// initialise common test parameters
	blocks = flag.Arg(0)
	period = flag.Arg(1)
	testname = flag.Arg(2)
	tempdir, err = ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}

	// set up worker pool
	wg := sync.WaitGroup{}
	for workerId := 0; workerId < jobs; workerId++ {
		wg.Add(1)

		go func(workerId int) {
			defer wg.Done()
			worker(workerId, queue)
		}(workerId)
	}

	// idiomatic hack required to use wg.Wait() with select
	waitCh := make(chan struct{})
	go func() {
		defer close(waitCh)
		wg.Wait()
	}()

wait:
	for {
		select {
		case <-waitCh:
			break wait
		case <-time.After(1 * time.Minute):
			fmt.Println(".")
		}
	}

	// analyse results and exit with 1 on first error
	close(results)
	for rc := range results {
		if !rc {
			os.Exit(1)
		}
	}

	os.Exit(0)
}

func buildCommand(testname, blocks, period, genesis string, seed int) string {
	return fmt.Sprintf("go test github.com/cosmos/cosmos-sdk/cmd/gaia/app -run %s -SimulationEnabled=true "+
		"-SimulationNumBlocks=%s -SimulationGenesis=%s "+
		"-SimulationVerbose=true -SimulationCommit=true -SimulationSeed=%d -SimulationPeriod=%s -v -timeout 24h",
		testname, blocks, genesis, seed, period)
}

func makeCmd(cmdStr string) *exec.Cmd {
	cmdSlice := strings.Split(cmdStr, " ")
	return exec.Command(cmdSlice[0], cmdSlice[1:]...)
}

func makeFilename(seed int) string {
	return fmt.Sprintf("gaia-simulation-seed-%d-date-%s", seed, time.Now().Format("01-02-2006_15:04:05.000000000"))
}

func worker(id int, seeds <-chan int) {
	log.Printf("[W%d] Worker is up and running", id)
	for seed := range seeds {
		if err := spawnProc(id, seed); err != nil {
			results <- false
			log.Printf("[W%d] Seed %d: FAILED", id, seed)
			log.Printf("To reproduce run: %s", buildCommand(testname, blocks, period, genesis, seed))
		} else {
			log.Printf("[W%d] Seed %d: OK", id, seed)
		}
	}
	log.Printf("[W%d] no seeds left, shutting down", id)
}

func spawnProc(workerId int, seed int) error {
	stderrFile, _ := os.Create(filepath.Join(tempdir, makeFilename(seed)+".stderr"))
	stdoutFile, _ := os.Create(filepath.Join(tempdir, makeFilename(seed)+".stdout"))
	s := buildCommand(testname, blocks, period, genesis, seed)
	cmd := makeCmd(s)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	err := cmd.Start()
	if err != nil {
		log.Printf("couldn't start %q", s)
		return err
	}
	log.Printf("[W%d] Spawned simulation with pid %d [seed=%d stdout=%s stderr=%s]",
		workerId, cmd.Process.Pid, seed, stdoutFile.Name(), stderrFile.Name())
	pushProcess(cmd.Process)
	defer popProcess(cmd.Process)
	return cmd.Wait()
}

func pushProcess(proc *os.Process) {
	mutex.Lock()
	defer mutex.Unlock()
	procs[proc.Pid] = proc
}

func popProcess(proc *os.Process) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := procs[proc.Pid]; ok {
		delete(procs, proc.Pid)
	}
}

func killAllProcs() {
	mutex.Lock()
	defer mutex.Unlock()
	for _, proc := range procs {
		checkSignal(proc, syscall.SIGTERM)
		checkSignal(proc, syscall.SIGKILL)
	}
}

func checkSignal(proc *os.Process, signal syscall.Signal) {
	if err := proc.Signal(signal); err != nil {
		log.Printf("Failed to send %s to PID %d", signal, proc.Pid)
	}
}
