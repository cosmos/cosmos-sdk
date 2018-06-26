package tests

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"time"
)

// execution process
type Process struct {
	ExecPath     string
	Args         []string
	Pid          int
	StartTime    time.Time
	EndTime      time.Time
	Cmd          *exec.Cmd        `json:"-"`
	ExitState    *os.ProcessState `json:"-"`
	WaitCh       chan struct{}    `json:"-"`
	StdinPipe    io.WriteCloser   `json:"-"`
	StdoutBuffer *bytes.Buffer    `json:"-"`
	StderrBuffer *bytes.Buffer    `json:"-"`
}

// dir: The working directory. If "", os.Getwd() is used.
// name: Command name
// args: Args to command. (should not include name)
// outFile, errFile: If not nil, will use, otherwise new Buffers will be
// allocated.  Either way, Process.Cmd.StdoutPipe and Process.Cmd.StderrPipe will be nil
// respectively.
func StartProcess(dir string, name string, args []string, outFile, errFile io.WriteCloser) (*Process, error) {
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
	// cmd stdout, stderr
	var outBuffer, errBuffer *bytes.Buffer
	if outFile != nil {
		cmd.Stdout = outFile
	} else {
		outBuffer = bytes.NewBuffer(nil)
		cmd.Stdout = outBuffer
	}
	if errFile != nil {
		cmd.Stderr = errFile
	} else {
		errBuffer = bytes.NewBuffer(nil)
		cmd.Stderr = errBuffer
	}
	// cmd start
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	proc := &Process{
		ExecPath:  name,
		Args:      args,
		Pid:       cmd.Process.Pid,
		StartTime: time.Now(),
		Cmd:       cmd,
		ExitState: nil,
		WaitCh:    make(chan struct{}),
		StdinPipe: stdin,
	}
	if outBuffer != nil {
		proc.StdoutBuffer = outBuffer
	}
	if errBuffer != nil {
		proc.StderrBuffer = errBuffer
	}
	go func() {
		err := proc.Cmd.Wait()
		if err != nil {
			// fmt.Printf("Process exit: %v\n", err)
			if exitError, ok := err.(*exec.ExitError); ok {
				proc.ExitState = exitError.ProcessState
			}
		}
		proc.ExitState = proc.Cmd.ProcessState
		proc.EndTime = time.Now() // TODO make this goroutine-safe
		close(proc.WaitCh)
	}()
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
	<-proc.WaitCh
}
