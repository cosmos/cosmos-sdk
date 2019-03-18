package init

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
)

func Test_prepareFlagsForTxCreateValidator(t *testing.T) {
	defer server.SetupViper(t)()
	defer setupClientHome(t)()
	config, err := tcmd.ParseConfig()
	require.Nil(t, err)
	logger := log.NewNopLogger()
	ctx := server.NewContext(config, logger)

	valPubKey, _ := sdk.GetConsPubKeyBech32("cosmosvalconspub1zcjduepq7jsrkl9fgqk0wj3ahmfr8pgxj6vakj2wzn656s8pehh0zhv2w5as5gd80a")

	type args struct {
		config    *cfg.Config
		nodeID    string
		ip        string
		chainID   string
		valPubKey crypto.PubKey
		website   string
		details   string
		identity  string
	}
	tests := []struct {
		name string
		args args
	}{
		{"No parameters", args{ctx.Config, "X", "0.0.0.0", "chainId", valPubKey, "", "", ""}},
		{"Oprtional parameters fed", args{ctx.Config, "X", "0.0.0.0", "chainId", valPubKey, "cosmos.network", "details", "identity"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepareFlagsForTxCreateValidator(tt.args.config, tt.args.nodeID, tt.args.ip, tt.args.chainID, tt.args.valPubKey, tt.args.website, tt.args.details, tt.args.identity)
			require.Equal(t, tt.args.website, viper.GetString(cli.FlagWebsite))
			require.Equal(t, tt.args.details, viper.GetString(cli.FlagDetails))
			require.Equal(t, tt.args.identity, viper.GetString(cli.FlagIdentity))
			require.Equal(t, defaultAmount, viper.GetString(cli.FlagAmount))
			require.Equal(t, defaultCommissionRate, viper.GetString(cli.FlagCommissionRate))
			require.Equal(t, defaultCommissionMaxRate, viper.GetString(cli.FlagCommissionMaxRate))
			require.Equal(t, defaultCommissionMaxChangeRate, viper.GetString(cli.FlagCommissionMaxChangeRate))
			require.Equal(t, defaultMinSelfDelegation, viper.GetString(cli.FlagMinSelfDelegation))
		})
	}

	var amount = "5stake"
	var commissionRate = "1.0"
	var commissionMaxRate = "1.0"
	var commissionMaxChangeRate = "1.0"
	var minSelfDelegation = "1.0"

	viper.Set(cli.FlagAmount, amount)
	viper.Set(cli.FlagCommissionRate, commissionRate)
	viper.Set(cli.FlagCommissionMaxRate, commissionMaxRate)
	viper.Set(cli.FlagCommissionMaxChangeRate, commissionMaxChangeRate)
	viper.Set(cli.FlagMinSelfDelegation, minSelfDelegation)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepareFlagsForTxCreateValidator(tt.args.config, tt.args.nodeID, tt.args.ip, tt.args.chainID, tt.args.valPubKey, tt.args.website, tt.args.details, tt.args.identity)
			require.Equal(t, tt.args.website, viper.GetString(cli.FlagWebsite))
			require.Equal(t, tt.args.details, viper.GetString(cli.FlagDetails))
			require.Equal(t, tt.args.identity, viper.GetString(cli.FlagIdentity))
			require.Equal(t, amount, viper.GetString(cli.FlagAmount))
			require.Equal(t, commissionRate, viper.GetString(cli.FlagCommissionRate))
			require.Equal(t, commissionMaxRate, viper.GetString(cli.FlagCommissionMaxRate))
			require.Equal(t, commissionMaxChangeRate, viper.GetString(cli.FlagCommissionMaxChangeRate))
			require.Equal(t, minSelfDelegation, viper.GetString(cli.FlagMinSelfDelegation))
		})
	}
}
