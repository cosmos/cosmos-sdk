package tendermint

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"

	yaml "gopkg.in/yaml.v2"
)

func TestEvienceString(t *testing.T) {
	dupEv := randomDuplicatedVoteEvidence()
	ev := Evidence{
		DuplicateVoteEvidence: dupEv,
		ChainID:               "mychain",
		ValidatorPower:        10,
		TotalPower:            50,
	}

	byteStr, err := yaml.Marshal(ev)
	require.Nil(t, err)
	require.Equal(t, string(byteStr), ev.String(), "Evidence String method does not work as expected")
}

func TestEvidenceValidateBasic(t *testing.T) {
	dupEv := randomDuplicatedVoteEvidence()

	testCases := []struct {
		message   string
		evidence  Evidence
		expectErr bool
	}{
		{
			"valid evidence",
			Evidence{
				DuplicateVoteEvidence: dupEv,
				ChainID:               "mychain",
				ValidatorPower:        10,
				TotalPower:            50,
			},
			false,
		},
		{
			"invalid duplicate vote evidence",
			Evidence{
				DuplicateVoteEvidence: &tmtypes.DuplicateVoteEvidence{
					PubKey: dupEv.PubKey,
					VoteA:  nil,
					VoteB:  dupEv.VoteB,
				},
				ChainID:        "mychain",
				ValidatorPower: 10,
				TotalPower:     50,
			},
			true,
		},
		{
			"empty duplicate vote evidence",
			Evidence{
				DuplicateVoteEvidence: nil,
				ChainID:               "mychain",
				ValidatorPower:        10,
				TotalPower:            50,
			},
			true,
		},
		{
			"empty chain ID",
			Evidence{
				DuplicateVoteEvidence: dupEv,
				ChainID:               "",
				ValidatorPower:        10,
				TotalPower:            50,
			},
			true,
		},
		{
			"invalid validator power",
			Evidence{
				DuplicateVoteEvidence: dupEv,
				ChainID:               "mychain",
				ValidatorPower:        0,
				TotalPower:            50,
			},
			true,
		},
		{
			"validator power > total power",
			Evidence{
				DuplicateVoteEvidence: dupEv,
				ChainID:               "mychain",
				ValidatorPower:        100,
				TotalPower:            50,
			},
			true,
		},
	}

	for i, tc := range testCases {

		require.Equal(t, tc.evidence.Route(), "client", "unexpected evidence route for tc #%d", i)
		require.Equal(t, tc.evidence.Type(), "client_misbehaviour", "unexpected evidence type for tc #%d", i)
		require.Equal(t, tc.evidence.GetValidatorPower(), tc.evidence.ValidatorPower, "unexpected val power for tc #%d", i)
		require.Equal(t, tc.evidence.GetTotalPower(), tc.evidence.TotalPower, "unexpected total power for tc #%d", i)
		require.Equal(t, tc.evidence.Hash(), cmn.HexBytes(tmhash.Sum(SubModuleCdc.MustMarshalBinaryBare(tc.evidence))), "unexpected evidence hash for tc #%d", i)

		if tc.expectErr {
			require.Error(t, tc.evidence.ValidateBasic(), "expected error for tc #%d", i)
		} else {
			require.Equal(t, tc.evidence.GetHeight(), tc.evidence.DuplicateVoteEvidence.Height(), "unexpected height for tc #%d", i)
			require.Equal(t, tc.evidence.GetConsensusAddress().String(), sdk.ConsAddress(tc.evidence.DuplicateVoteEvidence.Address()).String(), "unexpected cons addr for tc #%d", i)
			require.NoError(t, tc.evidence.ValidateBasic(), "unexpected error for tc #%d", i)
		}
	}
}
