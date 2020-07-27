package supervisor

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

// ExecuteT executes the command, pipes any input to STDIN and return STDOUT,
// logging STDOUT/STDERR to t.
func ExecuteT(t *testing.T, cmd, input string) (stdout, stderr string) {
	t.Log("Running", cmd)

	// split cmd to name and args
	split := strings.Split(cmd, " ")
	require.True(t, len(split) > 0, "no command provided")
	name, args := split[0], []string(nil)
	if len(split) > 1 {
		args = split[1:]
	}

	proc, err := StartProcess("", name, args)
	require.NoError(t, err)

	// if input is provided, pass it to STDIN and close the pipe
	if input != "" {
		_, err = io.WriteString(proc.StdinPipe, input)
		require.NoError(t, err)
		proc.StdinPipe.Close()
	}

	outbz, errbz, err := proc.ReadAll()
	if err != nil {
		fmt.Println("Err on proc.ReadAll()", err, args)
	}

	proc.Wait()

	if len(outbz) > 0 {
		t.Log("Stdout:", string(outbz))
	}

	if len(errbz) > 0 {
		t.Log("Stderr:", string(errbz))
	}

	stdout = strings.Trim(string(outbz), "\n")
	stderr = strings.Trim(string(errbz), "\n")

	return stdout, stderr
}

// Execute the command, launch goroutines to log stdout/err to t.
// Caller should wait for .Wait() or .Stop() to terminate.
func GoExecuteT(t *testing.T, cmd string) (proc *Process) {
	t.Log("Running", cmd)

	// Split cmd to name and args.
	split := strings.Split(cmd, " ")
	require.True(t, len(split) > 0, "no command provided")
	name, args := split[0], []string(nil)
	if len(split) > 1 {
		args = split[1:]
	}

	// Start process.
	proc, err := StartProcess("", name, args)
	require.NoError(t, err)
	return proc
}

// Same as GoExecuteT but spawns a go routine to ReadAll off stdout.
func GoExecuteTWithStdout(t *testing.T, cmd string) (proc *Process) {
	t.Log("Running", cmd)

	// Split cmd to name and args.
	split := strings.Split(cmd, " ")
	require.True(t, len(split) > 0, "no command provided")
	name, args := split[0], []string(nil)
	if len(split) > 1 {
		args = split[1:]
	}

	// Start process.
	proc, err := CreateProcess("", name, args)
	require.NoError(t, err)

	// Without this, the test halts ?!
	// (theory: because stdout and/or err aren't connected to anything the process halts)
	go func(proc *Process) {
		_, err := ioutil.ReadAll(proc.StdoutPipe)
		if err != nil {
			fmt.Println("-------------ERR-----------------------", err)
			return
		}
	}(proc)

	go func(proc *Process) {
		_, err := ioutil.ReadAll(proc.StderrPipe)
		if err != nil {
			fmt.Println("-------------ERR-----------------------", err)
			return
		}
	}(proc)

	err = proc.Cmd.Start()
	require.NoError(t, err)
	proc.Pid = proc.Cmd.Process.Pid
	return proc
}
