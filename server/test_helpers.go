package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tmlibs/cli"
)

// Get a free address for a test tendermint server
// protocol is either tcp, http, etc
func FreeTCPAddr() (addr, port string, err error) {
	l, err := net.Listen("tcp", "0.0.0.0:0")
	defer l.Close()
	if err != nil {
		return "", "", err
	}

	portI := l.Addr().(*net.TCPAddr).Port
	port = fmt.Sprintf("%d", portI)
	addr = fmt.Sprintf("tcp://0.0.0.0:%s", port)
	return
}

// setupViper creates a homedir to run inside,
// and returns a cleanup function to defer
func setupViper(t *testing.T) func() {
	rootDir, err := ioutil.TempDir("", "mock-sdk-cmd")
	require.Nil(t, err)
	viper.Set(cli.HomeFlag, rootDir)
	return func() {
		os.RemoveAll(rootDir)
	}
}

// Run or Timout RunE of command passed in
func RunOrTimeout(cmd *cobra.Command, timeout time.Duration, t *testing.T) chan error {
	done := make(chan error)
	go func(out chan<- error) {
		// this should NOT exit
		err := cmd.RunE(nil, nil)
		if err != nil {
			out <- err
		}
		out <- fmt.Errorf("start died for unknown reasons")
	}(done)
	timer := time.NewTimer(timeout)

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-timer.C:
		return done
	}
	return done
}
