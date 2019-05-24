package cli

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPrepareFlagsForTxCreateValidator(t *testing.T) {
	defer server.SetupViper(t)()
	config, err := tcmd.ParseConfig()
	require.Nil(t, err)
	logger := log.NewNopLogger()
	ctx := server.NewContext(config, logger)

	valPubKey, _ := sdk.GetConsPubKeyBech32("cosmosvalconspub1zcjduepq7jsrkl9fgqk0wj3ahmfr8pgxj6vakj2wzn656s8pehh0zhv2w5as5gd80a")

	type args struct {
		config    *cfg.Config
		nodeID    string
		chainID   string
		valPubKey crypto.PubKey
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
		PrepareFlagsForTxCreateValidator(tt.args.config, tt.args.nodeID,
			tt.args.chainID, tt.args.valPubKey)

		require.Equal(t, params.amount, viper.GetString(FlagAmount))
		require.Equal(t, params.commissionRate, viper.GetString(FlagCommissionRate))
		require.Equal(t, params.commissionMaxRate, viper.GetString(FlagCommissionMaxRate))
		require.Equal(t, params.commissionMaxChangeRate, viper.GetString(FlagCommissionMaxChangeRate))
		require.Equal(t, params.minSelfDelegation, viper.GetString(FlagMinSelfDelegation))
	}

	tests := []testcase{
		{"No parameters", args{ctx.Config, "X", "chainId", valPubKey}},
	}

	defaultParams := extraParams{
		defaultAmount,
		defaultCommissionRate,
		defaultCommissionMaxRate,
		defaultCommissionMaxChangeRate,
		defaultMinSelfDelegation,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) { runTest(t, tt, defaultParams) })
		})
	}

	// Override default params
	params := extraParams{"5stake", "1.0", "1.0", "1.0", "1.0"}
	viper.Set(FlagAmount, params.amount)
	viper.Set(FlagCommissionRate, params.commissionRate)
	viper.Set(FlagCommissionMaxRate, params.commissionMaxRate)
	viper.Set(FlagCommissionMaxChangeRate, params.commissionMaxChangeRate)
	viper.Set(FlagMinSelfDelegation, params.minSelfDelegation)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { runTest(t, tt, params) })
	}
}
