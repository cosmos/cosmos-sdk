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

	type extraParams struct {
		amount                  string
		commissionRate          string
		commissionMaxRate       string
		commissionMaxChangeRate string
		minSelfDelegation       string
	}

	type testcase struct {
		name string
		args args
	}

	runTest := func(t *testing.T, tt testcase, params extraParams) {
		prepareFlagsForTxCreateValidator(tt.args.config, tt.args.nodeID, tt.args.ip, tt.args.chainID, tt.args.valPubKey, tt.args.website, tt.args.details, tt.args.identity)
		require.Equal(t, tt.args.website, viper.GetString(cli.FlagWebsite))
		require.Equal(t, tt.args.details, viper.GetString(cli.FlagDetails))
		require.Equal(t, tt.args.identity, viper.GetString(cli.FlagIdentity))
		require.Equal(t, params.amount, viper.GetString(cli.FlagAmount))
		require.Equal(t, params.commissionRate, viper.GetString(cli.FlagCommissionRate))
		require.Equal(t, params.commissionMaxRate, viper.GetString(cli.FlagCommissionMaxRate))
		require.Equal(t, params.commissionMaxChangeRate, viper.GetString(cli.FlagCommissionMaxChangeRate))
		require.Equal(t, params.minSelfDelegation, viper.GetString(cli.FlagMinSelfDelegation))
	}

	tests := []testcase{
		{"No parameters", args{ctx.Config, "X", "0.0.0.0", "chainId", valPubKey, "", "", ""}},
		{"Optional parameters fed", args{ctx.Config, "X", "0.0.0.0", "chainId", valPubKey, "cosmos.network", "details", "identity"}},
	}

	defaultParams := extraParams{defaultAmount, defaultCommissionRate, defaultCommissionMaxRate, defaultCommissionMaxChangeRate, defaultMinSelfDelegation}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) { runTest(t, tt, defaultParams) })
		})
	}

	// Override default params
	params := extraParams{"5stake", "1.0", "1.0", "1.0", "1.0"}
	viper.Set(cli.FlagAmount, params.amount)
	viper.Set(cli.FlagCommissionRate, params.commissionRate)
	viper.Set(cli.FlagCommissionMaxRate, params.commissionMaxRate)
	viper.Set(cli.FlagCommissionMaxChangeRate, params.commissionMaxChangeRate)
	viper.Set(cli.FlagMinSelfDelegation, params.minSelfDelegation)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { runTest(t, tt, params) })
	}
}
