package gov_test

import (
	"bytes"
	"log"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject"
	sdklog "cosmossdk.io/log"
	"cosmossdk.io/math"
	_ "cosmossdk.io/x/auth"
	authtypes "cosmossdk.io/x/auth/types"
	_ "cosmossdk.io/x/bank"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"
	_ "cosmossdk.io/x/protocolpool"
	_ "cosmossdk.io/x/staking"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
)

var (
	valTokens           = sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	TestProposal        = v1beta1.NewTextProposal("Test", "description")
	TestDescription     = stakingtypes.NewDescription("T", "E", "S", "T", "Z")
	TestCommissionRates = stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
)

// mkTestLegacyContent creates a MsgExecLegacyContent for testing purposes.
func mkTestLegacyContent(t *testing.T) *v1.MsgExecLegacyContent {
	t.Helper()
	msgContent, err := v1.NewLegacyContent(TestProposal, authtypes.NewModuleAddress(types.ModuleName).String())
	assert.NilError(t, err)

	return msgContent
}

var pubkeys = []cryptotypes.PubKey{
	ed25519.GenPrivKey().PubKey(),
	ed25519.GenPrivKey().PubKey(),
	ed25519.GenPrivKey().PubKey(),
}

// SortAddresses - Sorts Addresses
func SortAddresses(addrs []sdk.AccAddress) {
	byteAddrs := make([][]byte, len(addrs))

	for i, addr := range addrs {
		byteAddrs[i] = addr.Bytes()
	}

	SortByteArrays(byteAddrs)

	for i, byteAddr := range byteAddrs {
		addrs[i] = byteAddr
	}
}

// implement `Interface` in sort package.
type sortByteArrays [][]byte

func (b sortByteArrays) Len() int {
	return len(b)
}

func (b sortByteArrays) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i], b[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (b sortByteArrays) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// SortByteArrays - sorts the provided byte array
func SortByteArrays(src [][]byte) [][]byte {
	sorted := sortByteArrays(src)
	sort.Sort(sorted)
	return sorted
}

func createTestSuite(t *testing.T) suite {
	t.Helper()
	res := suite{}

	app, err := simtestutil.SetupWithConfiguration(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.BankModule(),
				configurator.GovModule(),
				configurator.ConsensusModule(),
				configurator.ProtocolPoolModule(),
			),
			depinject.Supply(sdklog.NewNopLogger()),
		),
		simtestutil.DefaultStartUpConfig(),
		&res.AccountKeeper, &res.BankKeeper, &res.GovKeeper, &res.StakingKeeper,
	)
	require.NoError(t, err)

	res.app = app
	return res
}
