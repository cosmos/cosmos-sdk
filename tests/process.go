package tests

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

// execution process
type Process struct {
	ExecPath   string
	Args       []string
	Pid        int
	StartTime  time.Time
	EndTime    time.Time
	Cmd        *exec.Cmd        `json:"-"`
	ExitState  *os.ProcessState `json:"-"`
	StdinPipe  io.WriteCloser   `json:"-"`
	StdoutPipe io.ReadCloser    `json:"-"`
	StderrPipe io.ReadCloser    `json:"-"`
}

// dir: The working directory. If "", os.Getwd() is used.
// name: Command name
// args: Args to command. (should not include name)
func StartProcess(dir string, name string, args []string) (*Process, error) {
	proc, err := CreateProcess(dir, name, args)
	if err != nil {
		return nil, err
	}
	// cmd start
	if err := proc.Cmd.Start(); err != nil {
		return nil, err
	}
	proc.Pid = proc.Cmd.Process.Pid

	return proc, nil
}

// Same as StartProcess but doesn't start the process
func CreateProcess(dir string, name string, args []string) (*Process, error) {
	var cmd = exec.Command(name, args...) // is not yet started.
	// cmd dir
	if dir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		cmd.Dir = pwd
	} else {
		cmd.Dir = dir
	}
	// cmd stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	proc := &Process{
		ExecPath:   name,
		Args:       args,
		StartTime:  time.Now(),
		Cmd:        cmd,
		ExitState:  nil,
		StdinPipe:  stdin,
		StdoutPipe: stdout,
		StderrPipe: stderr,
	}
	return proc, nil
}

// stop the process
func (proc *Process) Stop(kill bool) error {
	if kill {
		// fmt.Printf("Killing process %v\n", proc.Cmd.Process)
		return proc.Cmd.Process.Kill()
	}
	return proc.Cmd.Process.Signal(os.Interrupt)
}

// wait for the process
func (proc *Process) Wait() {
	err := proc.Cmd.Wait()
	if err != nil {
		// fmt.Printf("Process exit: %v\n", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			proc.ExitState = exitError.ProcessState
		}
	}
	proc.ExitState = proc.Cmd.ProcessState
	proc.EndTime = time.Now() // TODO make this goroutine-safe
}

// ReadAll calls ioutil.ReadAll on the StdoutPipe and StderrPipe.
func (proc *Process) ReadAll() (stdout []byte, stderr []byte, err error) {
	outbz, err := ioutil.ReadAll(proc.StdoutPipe)
	if err != nil {
		return nil, nil, err
	}
	errbz, err := ioutil.ReadAll(proc.StderrPipe)
	if err != nil {
		return nil, nil, err
	}
	return outbz, errbz, nil
}
