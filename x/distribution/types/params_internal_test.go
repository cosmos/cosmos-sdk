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
	tests := []struct {
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
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantErr, validateCommunityTax(tt.args.i) != nil)
			require.Equal(t, tt.wantErr, validateBaseProposerReward(tt.args.i) != nil)
			require.Equal(t, tt.wantErr, validateBonusProposerReward(tt.args.i) != nil)
		})
	}
}
