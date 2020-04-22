package common

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

func TestQueryDelegationRewardsAddrValidation(t *testing.T) {
	viper.Set(flags.FlagTrustNode, true)
	ctx := context.NewCLIContext().WithCodec(codec.New())
	type args struct {
		delAddr string
		valAddr string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"invalid delegator address", args{"invalid", ""}, nil, true},
		{"empty delegator address", args{"", ""}, nil, true},
		{"invalid validator address", args{"cosmos1zxcsu7l5qxs53lvp0fqgd09a9r2g6kqrk2cdpa", "invalid"}, nil, true},
		{"empty validator address", args{"cosmos1zxcsu7l5qxs53lvp0fqgd09a9r2g6kqrk2cdpa", ""}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := QueryDelegationRewards(ctx, "", tt.args.delAddr, tt.args.valAddr)
			require.True(t, err != nil, tt.wantErr)
		})
	}
}
