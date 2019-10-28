package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"

	"github.com/stretchr/testify/require"
)

func TestMsgSubmitEvidence(t *testing.T) {
	submitter := sdk.AccAddress("test")
	testCases := []struct {
		evidence  types.Evidence
		submitter sdk.AccAddress
		expectErr bool
	}{
		{nil, submitter, true},
		// TODO: Add test cases using real concrete types.
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
