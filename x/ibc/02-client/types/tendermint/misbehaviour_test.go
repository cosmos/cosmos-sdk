package tendermint

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

func TestMisbehaviour(t *testing.T) {
	testCases := []struct {
		message      string
		misbehaviour Misbehaviour
		expectErr    bool
	}{
		{
			"valid misbehaviour",
			Misbehaviour{
				Evidence: &Evidence{
					DuplicateVoteEvidence: randomDuplicatedVoteEvidence(),
					ChainID:               "mychain",
					ValidatorPower:        10,
					TotalPower:            50,
				},
				ClientID: "ibcclientzero",
			},
			false,
		},
		{
			"empty misbehaviour evidence",
			Misbehaviour{
				Evidence: nil,
				ClientID: "ibcclientzero",
			},
			true,
		},
		{
			"empty misbehaviour evidence",
			Misbehaviour{
				Evidence: &Evidence{
					DuplicateVoteEvidence: randomDuplicatedVoteEvidence(),
					ChainID:               "mychain",
					ValidatorPower:        100,
					TotalPower:            50,
				},
				ClientID: "ibcclientzero",
			},
			true,
		},
		{
			"invalid client ID",
			Misbehaviour{
				Evidence: &Evidence{
					DuplicateVoteEvidence: randomDuplicatedVoteEvidence(),
					ChainID:               "mychain",
					ValidatorPower:        10,
					TotalPower:            50,
				},
				ClientID: " ",
			},
			true,
		},
	}

	for i, tc := range testCases {
		require.Equal(t, tc.misbehaviour.ClientType().String(), exported.ClientTypeTendermint, "unexpected misbehaviour client type for tc #%d", i)
		require.Equal(t, tc.misbehaviour.GetEvidence(), tc.misbehaviour.Evidence, "unexpected evidence for tc #%d", i)

		if tc.expectErr {
			require.Error(t, tc.misbehaviour.ValidateBasic(), "expected error for tc #%d", i)
		} else {
			require.NoError(t, tc.misbehaviour.ValidateBasic(), "unexpected error for tc #%d", i)
			require.NoError(t, CheckMisbehaviour(tc.misbehaviour))
		}
	}
}
