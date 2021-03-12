package upgrade_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
)

func TestInitChainer(t *testing.T) {
	app := simapp.Setup(false)

	app.InitChain(
		abcitypes.RequestInitChain{
			AppStateBytes: []byte("{}"),
			ChainId:       "test-chain-id",
		},
	)

	versionMap := app.GetVersionMap()
	require.NotEmpty(t, versionMap)
}
