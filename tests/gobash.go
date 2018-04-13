package tests

import (
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

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
func ExecuteT(t *testing.T, command string) (out string) {
	cmd := getCmd(t, command)
	bz, err := cmd.CombinedOutput()
	require.NoError(t, err, string(bz))
	out = strings.Trim(string(bz), "\n") //trim any new lines
	time.Sleep(time.Second)
	return out
}

// Asynchronously execute the command, return standard output and error
func GoExecuteT(t *testing.T, command string) (cmd *exec.Cmd, pipeIn io.WriteCloser, pipeOut io.ReadCloser) {
	cmd = getCmd(t, command)
	pipeIn, err := cmd.StdinPipe()
	require.NoError(t, err)
	pipeOut, err = cmd.StdoutPipe()
	require.NoError(t, err)

	go func() {
		cmd.Start()
		//bz, _ := cmd.CombinedOutput()
		//require.NoError(t, err, string(bz))
		//outChan <- strings.Trim(string(bz), "\n") //trim any new lines
	}()
	time.Sleep(time.Second)
	return cmd, pipeIn, pipeOut
}
