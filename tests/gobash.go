package tests

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// Execute the command, return stdout, logging stdout/err to t.
func ExecuteT(t *testing.T, cmd string) (out string) {
	t.Log("Running", cmn.Cyan(cmd))

	// Split cmd to name and args.
	split := strings.Split(cmd, " ")
	require.True(t, len(split) > 0, "no command provided")
	name, args := split[0], []string(nil)
	if len(split) > 1 {
		args = split[1:]
	}

	// Start process and wait.
	proc, err := StartProcess("", name, args)
	require.NoError(t, err)

	// Get the output.
	outbz, errbz, err := proc.ReadAll()
	if err != nil {
		fmt.Println("Err on proc.ReadAll()", err, args)
	}
	proc.Wait()

	// Log output.
	if len(outbz) > 0 {
		t.Log("Stdout:", cmn.Green(string(outbz)))
	}
	if len(errbz) > 0 {
		t.Log("Stderr:", cmn.Red(string(errbz)))
	}

	// Collect STDOUT output.
	out = strings.Trim(string(outbz), "\n") //trim any new lines
	return out
}

// Execute the command, launch goroutines to log stdout/err to t.
// Caller should wait for .Wait() or .Stop() to terminate.
func GoExecuteT(t *testing.T, cmd string) (proc *Process) {
	t.Log("Running", cmn.Cyan(cmd))

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
	t.Log("Running", cmn.Cyan(cmd))

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
	go func() {
		_, err := ioutil.ReadAll(proc.StdoutPipe)
		if err != nil {
			fmt.Println("-------------ERR-----------------------", err)
			return
		}
	}()

	err = proc.Cmd.Start()
	require.NoError(t, err)
	proc.Pid = proc.Cmd.Process.Pid
	return proc
}
