package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestMsgSubmitEvidence(t *testing.T) {
	pk := ed25519.GenPrivKey()
	sv := types.TestVote{
		ValidatorAddress: pk.PubKey().Address(),
		Height:           11,
		Round:            0,
	}
	sig, err := pk.Sign(sv.SignBytes("test-chain"))
	require.NoError(t, err)
	sv.Signature = sig

	submitter := sdk.AccAddress("test")
	testCases := []struct {
		evidence  exported.Evidence
		submitter sdk.AccAddress
		expectErr bool
	}{
		{nil, submitter, true},
		{
			types.TestEquivocationEvidence{
				Power:      100,
				TotalPower: 100000,
				PubKey:     pk.PubKey(),
				VoteA:      sv,
				VoteB:      sv,
			},
			submitter,
			false,
		},
		{
			types.TestEquivocationEvidence{
				Power:      100,
				TotalPower: 100000,
				PubKey:     pk.PubKey(),
				VoteA:      sv,
				VoteB:      types.TestVote{Height: 10, Round: 1},
			},
			submitter,
			true,
		},
	}

	for i, tc := range testCases {
		msg := types.NewMsgSubmitEvidence(tc.evidence, tc.submitter)
		require.Equal(t, msg.Route(), types.RouterKey, "unexpected result for tc #%d", i)
		require.Equal(t, msg.Type(), types.TypeMsgSubmitEvidence, "unexpected result for tc #%d", i)
		require.Equal(t, tc.expectErr, msg.ValidateBasic() != nil, "unexpected result for tc #%d", i)

		if !tc.expectErr {
			require.Equal(t, msg.GetSigners(), []sdk.AccAddress{tc.submitter}, "unexpected result for tc #%d", i)
		}
	}
}
