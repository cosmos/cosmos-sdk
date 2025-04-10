package cmtservice_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/server"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

func TestStatusCommand(t *testing.T) {
	t.Skip() // flaky test

	cfg, err := network.DefaultConfigWithAppConfig(network.MinimumAppConfig())
	require.NoError(t, err)

	network, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	require.NoError(t, network.WaitForNextBlock())

	val0 := network.Validators[0]
	cmd := server.StatusCommand()

	out, err := clitestutil.ExecTestCLICmd(val0.ClientCtx, cmd, []string{})
	require.NoError(t, err)

	// Make sure the output has the validator moniker.
	require.Contains(t, out.String(), fmt.Sprintf("\"moniker\":\"%s\"", val0.Moniker))
}
