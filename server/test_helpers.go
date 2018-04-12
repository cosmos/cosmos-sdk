package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/mock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tmlibs/cli"
	"github.com/tendermint/tmlibs/log"
)

// Get a free address for a test tendermint server
// protocol is either tcp, http, etc
func FreeTCPAddr(t *testing.T) string {
	l, err := net.Listen("tcp", "0.0.0.0:0")
	defer l.Close()
	require.Nil(t, err)

	port := l.Addr().(*net.TCPAddr).Port
	addr := fmt.Sprintf("tcp://0.0.0.0:%d", port)
	return addr
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

// Begin the server pass up the channel to close
// NOTE pass up the channel so it can be closed at the end of the process
func StartServer(t *testing.T) chan error {
	defer setupViper(t)()

	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)

	// init server
	ctx := NewContext(cfg, log.NewNopLogger())
	initCmd := InitCmd(mock.GenInitOptions, ctx)
	err = initCmd.RunE(nil, nil)
	require.NoError(t, err)

	// start server
	viper.Set(flagWithTendermint, true)
	startCmd := StartCmd(mock.NewApp, ctx)
	startCmd.Flags().Set(flagAddress, FreeTCPAddr(t)) // set to a new free address
	startCmd.Flags().Set("rpc.laddr", FreeTCPAddr(t)) // set to a new free address
	timeout := time.Duration(3) * time.Second

	return RunOrTimeout(startCmd, timeout, t)
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
