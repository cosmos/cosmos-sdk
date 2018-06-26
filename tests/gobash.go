package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	cmn "github.com/tendermint/tmlibs/common"
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
	proc, err := StartProcess("", name, args, nil, nil)
	require.NoError(t, err)
	proc.Wait()

	// Get the output.
	outbz := proc.StdoutBuffer.Bytes()
	errbz := proc.StderrBuffer.Bytes()

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
	proc, err := StartProcess("", name, args, nil, nil)
	require.NoError(t, err)

	// Run goroutines to log stdout.
	go func() {
		buf := make([]byte, 10240) // TODO Document the effects.
		for {
			n, err := proc.StdoutBuffer.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				t.Log("Stdout:", cmn.Green(string(buf[:n])))
			}
		}
	}()

	// Run goroutines to log stderr.
	go func() {
		buf := make([]byte, 10240) // TODO Document the effects.
		for {
			n, err := proc.StderrBuffer.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				t.Log("Stderr:", cmn.Red(string(buf[:n])))
			}
		}
	}()

	return proc
}
