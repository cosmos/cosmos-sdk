package testnet_test

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	cmtcfg "github.com/cometbft/cometbft/config"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testnet"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Use a limited set of available ports to ensure that
// retries eventually land on a free port.
func TestCometStarter_PortContention(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long test in short mode")
	}

	const nVals = 4

	// Find n+1 addresses that should be free.
	// Ephemeral port range should start at about 49k+
	// according to `sysctl net.inet.ip.portrange` on macOS,
	// and at about 32k+ on Linux
	// according to `sysctl net.ipv4.ip_local_port_range`.
	//
	// Because we attempt to find free addresses outside that range,
	// it is unlikely that another process will claim a port
	// we discover to be free, during the time this test runs.
	const portSeekStart = 19000
	reuseAddrs := make([]string, 0, nVals+1)
	for i := portSeekStart; i < portSeekStart+1000; i++ {
		addr := fmt.Sprintf("127.0.0.1:%d", i)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			// No need to log the failure.
			continue
		}

		// If the port was free, append it to our reusable addresses.
		reuseAddrs = append(reuseAddrs, "tcp://"+addr)
		_ = ln.Close()

		if len(reuseAddrs) == nVals+1 {
			break
		}
	}

	if len(reuseAddrs) != nVals+1 {
		t.Fatalf("needed %d free ports but only found %d", nVals+1, len(reuseAddrs))
	}

	// Now that we have one more port than the number of validators,
	// there is a good chance that picking a random port will conflict with a previously chosen one.
	// But since CometStarter retries several times,
	// it should eventually land on a free port.

	valPKs := testnet.NewValidatorPrivKeys(nVals)
	cmtVals := valPKs.CometGenesisValidators()
	stakingVals := cmtVals.StakingValidators()

	const chainID = "simapp-cometstarter"

	b := testnet.DefaultGenesisBuilderOnlyValidators(
		chainID,
		stakingVals,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction),
	)

	jGenesis := b.Encode()

	// Use an info-level logger, because the debug logs in comet are noisy
	// and there is a data race in comet debug logs,
	// due to be fixed in v0.37.1 which is not yet released:
	// https://github.com/cometbft/cometbft/pull/532
	logger := log.NewTestLoggerInfo(t)

	const nRuns = 4
	for i := 0; i < nRuns; i++ {
		t.Run(fmt.Sprintf("attempt %d", i), func(t *testing.T) {
			nodes, err := testnet.NewNetwork(nVals, func(idx int) *testnet.CometStarter {
				rootDir := t.TempDir()

				app := simapp.NewSimApp(
					logger.With("instance", idx),
					dbm.NewMemDB(),
					nil,
					true,
					simtestutil.NewAppOptionsWithFlagHome(rootDir),
					baseapp.SetChainID(chainID),
				)

				cfg := cmtcfg.DefaultConfig()

				// memdb is sufficient for this test.
				cfg.BaseConfig.DBBackend = "memdb"

				return testnet.NewCometStarter(
					app,
					cfg,
					valPKs[idx].Val,
					jGenesis,
					rootDir,
				).
					Logger(logger.With("rootmodule", fmt.Sprintf("comet_node-%d", idx))).
					TCPAddrChooser(func() string {
						// This chooser function is the key of this test,
						// where there is only one more available address than there are nodes.
						// Therefore it is likely that an address will already be in use,
						// thereby exercising the address-in-use retry.
						return reuseAddrs[rand.Intn(len(reuseAddrs))]
					})
			})
			defer nodes.StopAndWait()
			require.NoError(t, err)

			heightAdvanced := false
			for j := 0; j < 40; j++ {
				cs := nodes[0].ConsensusState()
				if cs.GetLastHeight() < 2 {
					time.Sleep(250 * time.Millisecond)
					continue
				}

				// Saw height advance.
				heightAdvanced = true
				break
			}

			if !heightAdvanced {
				t.Fatalf("consensus height did not advance in approximately 10 seconds")
			}
		})
	}
}
