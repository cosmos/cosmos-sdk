package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_validateAuxFuncs(t *testing.T) {
	type args struct {
		i interface{}
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"wrong type", args{10.5}, true},
		{"nil Int pointer", args{sdk.Dec{}}, true},
		{"negative", args{sdk.NewDec(-1)}, true},
		{"one dec", args{sdk.NewDec(1)}, false},
		{"two dec", args{sdk.NewDec(2)}, true},
	}

	for _, tc := range testCases {
		stc := tc

		t.Run(stc.name, func(t *testing.T) {
			require.Equal(t, stc.wantErr, validateCommunityTax(stc.args.i) != nil)
			require.Equal(t, stc.wantErr, validateBaseProposerReward(stc.args.i) != nil)
			require.Equal(t, stc.wantErr, validateBonusProposerReward(stc.args.i) != nil)
		})
	}
}
