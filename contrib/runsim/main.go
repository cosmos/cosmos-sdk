package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	GithubConfigSep = ","
	SlackConfigSep  = ","
)

var (
	// default seeds
	seeds = []int{
		1, 2, 4, 7, 32, 123, 124, 582, 1893, 2989,
		3012, 4728, 37827, 981928, 87821, 891823782,
		989182, 89182391, 11, 22, 44, 77, 99, 2020,
		3232, 123123, 124124, 582582, 18931893,
		29892989, 30123012, 47284728, 7601778, 8090485,
		977367484, 491163361, 424254581, 673398983,
	}
	seedOverrideList = ""

	// goroutine-safe process map
	procs map[int]*os.Process
	mutex *sync.Mutex

	// results channel
	results chan bool

	// command line arguments and options
	jobs         = runtime.GOMAXPROCS(0)
	pkgName      string
	blocks       string
	period       string
	testname     string
	genesis      string
	exitOnFail   bool
	githubConfig string
	gitRevision  string
	slackConfig  string

	// integration with Slack and Github
	slackToken   string
	slackChannel string
	slackThread  string

	// logs temporary directory
	tempdir string
)

func init() {
	log.SetPrefix("")
	log.SetFlags(0)

	runsimLogfile, err := os.OpenFile("sim_log_file", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("ERROR: opening log file: %v", err.Error())
	} else {
		log.SetOutput(io.MultiWriter(os.Stdout, runsimLogfile))
	}

	procs = map[int]*os.Process{}
	mutex = &sync.Mutex{}

	flag.IntVar(&jobs, "j", jobs, "Number of parallel processes")
	flag.StringVar(&genesis, "g", "", "Genesis file")
	flag.StringVar(&seedOverrideList, "seeds", "", "run the supplied comma-separated list of seeds instead of defaults")
	flag.BoolVar(&exitOnFail, "e", false, "Exit on fail during multi-sim, print error")
	flag.StringVar(&gitRevision, "rev", "", "git revision")
	flag.StringVar(&githubConfig, "github", "", "Report results to Github's PR")
	flag.StringVar(&slackConfig, "slack", "", "Report results to slack channel")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [-j maxprocs] [-seeds comma-separated-seed-list] [-rev git-commmit-hash] [-g genesis.json] [-e] [-github token,pr-url] [-slack token,channel,thread] [package] [blocks] [period] [testname]
Run simulations in parallel`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func main() {
	var err error

	flag.Parse()
	if flag.NArg() != 4 {
		log.Fatal("wrong number of arguments")
	}

	if githubConfig != "" {
		opts := strings.Split(githubConfig, GithubConfigSep)
		if len(opts) != 2 {
			log.Fatal("incorrect github config string format")
		}
	}

	if slackConfig != "" {
		opts := strings.Split(slackConfig, SlackConfigSep)
		if len(opts) != 3 {
			log.Fatal("incorrect slack config string format")
		}
		slackToken, slackChannel, slackThread = opts[0], opts[1], opts[2]
	}

	seedOverrideList = strings.TrimSpace(seedOverrideList)
	if seedOverrideList != "" {
		seeds, err = makeSeedList(seedOverrideList)
		if err != nil {
			log.Fatal(err)
		}
	}

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
	pkgName = flag.Arg(0)
	blocks = flag.Arg(1)
	period = flag.Arg(2)
	testname = flag.Arg(3)
	tempdir, err = ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}

	// set up worker pool
	log.Printf("Allocating %d workers...", jobs)
	wg := sync.WaitGroup{}
	for workerID := 0; workerID < jobs; workerID++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			worker(workerID, queue)
		}(workerID)
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
	if slackConfigSupplied() {
		seedStrings := make([]string, len(seeds))
		for i, seed := range seeds {
			seedStrings[i] = fmt.Sprintf("%d", seed)
		}
		slackMessage(slackToken, slackChannel, &slackThread, fmt.Sprintf("Finished running simulation for seeds: %s", strings.Join(seedStrings, " ")))
	}
	os.Exit(0)
}

func buildCommand(testName, blocks, period, genesis string, seed int) string {
	return fmt.Sprintf("go test %s -run %s -SimulationEnabled=true "+
		"-SimulationNumBlocks=%s -SimulationGenesis=%s "+
		"-SimulationVerbose=true -SimulationCommit=true -SimulationSeed=%d -SimulationPeriod=%s -v -timeout 24h",
		pkgName, testName, blocks, genesis, seed, period)
}

func makeCmd(cmdStr string) *exec.Cmd {
	cmdSlice := strings.Split(cmdStr, " ")
	return exec.Command(cmdSlice[0], cmdSlice[1:]...)
}

func makeFilename(seed int) string {
	return fmt.Sprintf("app-simulation-seed-%d-date-%s", seed, time.Now().Format("01-02-2006_15:04:05.000000000"))
}

func worker(id int, seeds <-chan int) {
	log.Printf("[W%d] Worker is up and running", id)
	for seed := range seeds {
		stdOut, stdErr, err := spawnProc(id, seed)
		if err != nil {
			results <- false
			log.Printf("[W%d] Seed %d: FAILED", id, seed)
			log.Printf("To reproduce run: %s", buildCommand(testname, blocks, period, genesis, seed))
			if slackConfigSupplied() {
				slackMessage(slackToken, slackChannel, nil, "Seed "+strconv.Itoa(seed)+" failed. To reproduce, run: "+buildCommand(testname, blocks, period, genesis, seed))
			}
			if exitOnFail {
				log.Printf("\bERROR OUTPUT \n\n%s", err)
				panic("halting simulations")
			}
		} else {
			log.Printf("[W%d] Seed %d: OK", id, seed)
		}
		pushLogs(stdOut, stdErr, gitRevision)
	}

	log.Printf("[W%d] no seeds left, shutting down", id)
}

func spawnProc(workerID int, seed int) (*os.File, *os.File, error) {
	stderrFile, _ := os.Create(filepath.Join(tempdir, makeFilename(seed)+".stderr"))
	stdoutFile, _ := os.Create(filepath.Join(tempdir, makeFilename(seed)+".stdout"))
	s := buildCommand(testname, blocks, period, genesis, seed)
	cmd := makeCmd(s)
	cmd.Stdout = stdoutFile

	var err error
	var stderr io.ReadCloser
	if !exitOnFail {
		cmd.Stderr = stderrFile
	} else {
		stderr, err = cmd.StderrPipe()
		if err != nil {
			return nil, nil, err
		}
	}
	sc := bufio.NewScanner(stderr)

	err = cmd.Start()
	if err != nil {
		log.Printf("couldn't start %q", s)
		return nil, nil, err
	}
	log.Printf("[W%d] Spawned simulation with pid %d [seed=%d stdout=%s stderr=%s]",
		workerID, cmd.Process.Pid, seed, stdoutFile.Name(), stderrFile.Name())
	pushProcess(cmd.Process)
	defer popProcess(cmd.Process)

	err = cmd.Wait()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	if exitOnFail {
		for sc.Scan() {
			fmt.Printf("stderr: %s\n", sc.Text())
		}
	}
	return stdoutFile, stderrFile, err
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

func makeSeedList(seeds string) ([]int, error) {
	strSeedsLst := strings.Split(seeds, ",")
	if len(strSeedsLst) == 0 {
		return nil, fmt.Errorf("seeds was empty")
	}
	intSeeds := make([]int, len(strSeedsLst))
	for i, seedstr := range strSeedsLst {
		intSeed, err := strconv.Atoi(strings.TrimSpace(seedstr))
		if err != nil {
			return nil, fmt.Errorf("cannot convert seed to integer: %v", err)
		}
		intSeeds[i] = intSeed
	}
	return intSeeds, nil
}

func slackConfigSupplied() bool { return slackConfig != "" }
