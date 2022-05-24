package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestParams_ValidateBasic(t *testing.T) {
	toDec := sdk.MustNewDecFromStr

	type fields struct {
		CommunityTax        sdk.Dec
		BaseProposerReward  sdk.Dec
		BonusProposerReward sdk.Dec
		WithdrawAddrEnabled bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"success", fields{toDec("0.1"), toDec("0.5"), toDec("0.4"), false}, false},
		{"negative community tax", fields{toDec("-0.1"), toDec("0.5"), toDec("0.4"), false}, true},
		{"negative base proposer reward", fields{toDec("0.1"), toDec("-0.5"), toDec("0.4"), false}, true},
		{"negative bonus proposer reward", fields{toDec("0.1"), toDec("0.5"), toDec("-0.4"), false}, true},
		{"total sum greater than 1", fields{toDec("0.2"), toDec("0.5"), toDec("0.4"), false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := types.Params{
				CommunityTax:        tt.fields.CommunityTax,
				BaseProposerReward:  tt.fields.BaseProposerReward,
				BonusProposerReward: tt.fields.BonusProposerReward,
				WithdrawAddrEnabled: tt.fields.WithdrawAddrEnabled,
			}
			if err := p.ValidateBasic(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	require.NoError(t, types.DefaultParams().ValidateBasic())
}
