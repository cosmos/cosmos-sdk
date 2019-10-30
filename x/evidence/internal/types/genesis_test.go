package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestDefaultGenesisState(t *testing.T) {
	gs := types.DefaultGenesisState()
	require.NotNil(t, gs.Evidence)
	require.Len(t, gs.Evidence, 0)
}

func TestGenesisStateValidate_Valid(t *testing.T) {
	pk := ed25519.GenPrivKey()

	evidence := make([]types.Evidence, 100)
	for i := 0; i < 100; i++ {
		sv := types.TestVote{
			ValidatorAddress: pk.PubKey().Address(),
			Height:           int64(i),
			Round:            0,
		}
		sig, err := pk.Sign(sv.SignBytes("test-chain"))
		require.NoError(t, err)
		sv.Signature = sig

		evidence[i] = types.TestEquivocationEvidence{
			Power:      100,
			TotalPower: 100000,
			PubKey:     pk.PubKey(),
			VoteA:      sv,
			VoteB:      sv,
		}
	}

	gs := types.NewGenesisState(evidence)
	require.NoError(t, gs.Validate())
}

func TestGenesisStateValidate_Invalid(t *testing.T) {
	pk := ed25519.GenPrivKey()

	evidence := make([]types.Evidence, 100)
	for i := 0; i < 100; i++ {
		sv := types.TestVote{
			ValidatorAddress: pk.PubKey().Address(),
			Height:           int64(i),
			Round:            0,
		}
		sig, err := pk.Sign(sv.SignBytes("test-chain"))
		require.NoError(t, err)
		sv.Signature = sig

		evidence[i] = types.TestEquivocationEvidence{
			Power:      100,
			TotalPower: 100000,
			PubKey:     pk.PubKey(),
			VoteA:      sv,
			VoteB:      types.TestVote{Height: 10, Round: 1},
		}
	}

	gs := types.NewGenesisState(evidence)
	require.Error(t, gs.Validate())
}
