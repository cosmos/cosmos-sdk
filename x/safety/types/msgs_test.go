package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

func TestMsgsValidateBasic(t *testing.T) {
	update := func(scope ...string) *types.MsgUpdateParams {
		return &types.MsgUpdateParams{Params: types.Params{Scope: scope}}
	}
	for name, tc := range map[string]struct {
		msg     interface{ ValidateBasic() error }
		wantErr error
	}{
		"update: empty scope":       {update(), nil},
		"update: valid scope":       {update("/cosmos.bank.*", "/cosmos.staking.v1beta1.MsgDelegate"), nil},
		"update: malformed scope":   {update("cosmos.bank.*"), types.ErrInvalidPattern},
		"update: bad segment":       {update("/cosmos.bank-x.*"), types.ErrInvalidPattern},
		"update: duplicate scope":   {update("/cosmos.bank.*", "/cosmos.bank.*"), types.ErrDuplicatePattern},
		"trip: valid":               {types.NewMsgTrip("", "/cosmos.bank.*"), nil},
		"trip: no patterns":         {types.NewMsgTrip(""), types.ErrEmptyPatterns},
		"trip: malformed pattern":   {types.NewMsgTrip("", "/cosmos..bank"), types.ErrInvalidPattern},
		"reset: valid":              {types.NewMsgReset("", "*"), nil},
		"reset: malformed pattern":  {types.NewMsgReset("", "**"), types.ErrInvalidPattern},
		"reset: duplicate patterns": {types.NewMsgReset("", "*", "*"), types.ErrDuplicatePattern},
	} {
		t.Run(name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
