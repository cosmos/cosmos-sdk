package tests

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ExecuteT executes the command, pipes any input to STDIN and return STDOUT,
// logging STDOUT/STDERR to t.
// nolint: errcheck
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

	return
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
