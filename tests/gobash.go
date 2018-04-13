package tests

import (
	"io"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func getCmd(t *testing.T, command string) *exec.Cmd {

	//split command into command and args
	split := strings.Split(command, " ")
	require.True(t, len(split) > 0, "no command provided")

	var cmd *exec.Cmd
	if len(split) == 1 {
		cmd = exec.Command(split[0])
	} else {
		cmd = exec.Command(split[0], split[1:]...)
	}
	return cmd
}

// Execute the command, return standard output and error
func ExecuteT(t *testing.T, command string) (pipe io.WriteCloser, out string) {
	cmd := getCmd(t, command)
	pipe, err := cmd.StdinPipe()
	require.NoError(t, err)
	bz, err := cmd.CombinedOutput()
	require.NoError(t, err)
	out = strings.Trim(string(bz), "\n") //trim any new lines
	return pipe, out
}

// Asynchronously execute the command, return standard output and error
func GoExecuteT(t *testing.T, command string) (pipe io.WriteCloser, outChan chan string) {
	cmd := getCmd(t, command)
	pipe, err := cmd.StdinPipe()
	require.NoError(t, err)
	go func() {
		bz, err := cmd.CombinedOutput()
		require.NoError(t, err)
		outChan <- strings.Trim(string(bz), "\n") //trim any new lines
	}()
	return pipe, outChan
}
