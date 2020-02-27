package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

func TestDefaultGenesisState(t *testing.T) {
	gs := types.DefaultGenesisState()
	require.NotNil(t, gs.Evidence)
	require.Len(t, gs.Evidence, 0)
}

func TestGenesisStateValidate_Valid(t *testing.T) {
	pk := ed25519.GenPrivKey()

	evidence := make([]exported.Evidence, 100)
	for i := 0; i < 100; i++ {
		evidence[i] = types.Equivocation{
			Height:           int64(i) + 1,
			Power:            100,
			Time:             time.Now().UTC(),
			ConsensusAddress: pk.PubKey().Address().Bytes(),
		}
	}

	gs := types.NewGenesisState(types.DefaultParams(), evidence)
	require.NoError(t, gs.Validate())
}

func TestGenesisStateValidate_Invalid(t *testing.T) {
	pk := ed25519.GenPrivKey()

	evidence := make([]exported.Evidence, 100)
	for i := 0; i < 100; i++ {
		evidence[i] = types.Equivocation{
			Height:           int64(i),
			Power:            100,
			Time:             time.Now().UTC(),
			ConsensusAddress: pk.PubKey().Address().Bytes(),
		}
	}

	gs := types.NewGenesisState(types.DefaultParams(), evidence)
	require.Error(t, gs.Validate())
}
