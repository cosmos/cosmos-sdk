package cli

import (
	"testing"

	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPrepareConfigForTxCreateValidator(t *testing.T) {
	chainID := "chainID"
	ip := "1.1.1.1"
	nodeID := "nodeID"
	valPubKey, _ := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, "cosmosvalconspub1zcjduepq7jsrkl9fgqk0wj3ahmfr8pgxj6vakj2wzn656s8pehh0zhv2w5as5gd80a")
	moniker := "myMoniker"

	tests := []struct {
		name        string
		config      func() *cfg.Config
		fsModify    func(fs *pflag.FlagSet)
		expectedCfg TxCreateValidatorConfig
	}{
		{
			name: "all defaults",
			config: func() *cfg.Config {
				config := &cfg.Config{BaseConfig: cfg.TestBaseConfig()}
				config.Moniker = moniker
				return config
			},
			fsModify: func(fs *pflag.FlagSet) {
				return
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                ip,
				ChainID:           chainID,
				NodeID:            nodeID,
				TrustNode:         true,
				PubKey:            sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Moniker:           moniker,
				Amount:            "100000000stake",
				CommissionRate:    "0.1",
				CommissionMaxRate: "0.2",
			},
		},
		{
			name: "If moniker is empty it sets from Flag.",
			config: func() *cfg.Config {
				config := &cfg.Config{BaseConfig: cfg.TestBaseConfig()}
				config.Moniker = ""

				return config
			},
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(flags.FlagName, "theNameFlag")
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                ip,
				From:              "theNameFlag",
				Moniker:           "theNameFlag",
				ChainID:           chainID,
				NodeID:            nodeID,
				TrustNode:         true,
				PubKey:            sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Amount:            "100000000stake",
				CommissionRate:    "0.1",
				CommissionMaxRate: "0.2",
			},
		},
		{
			name: "Custom amount",
			config: func() *cfg.Config {
				config := &cfg.Config{BaseConfig: cfg.TestBaseConfig()}
				config.Moniker = ""

				return config
			},
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(flags.FlagName, "theNameFlag")
				fs.Set(FlagAmount, "2000stake")
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                ip,
				From:              "theNameFlag",
				Moniker:           "theNameFlag",
				ChainID:           chainID,
				NodeID:            nodeID,
				TrustNode:         true,
				PubKey:            sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Amount:            "2000stake",
				CommissionRate:    "0.1",
				CommissionMaxRate: "0.2",
			},
		},
		{
			name: "Custom commission rate",
			config: func() *cfg.Config {
				config := &cfg.Config{BaseConfig: cfg.TestBaseConfig()}
				config.Moniker = ""

				return config
			},
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(flags.FlagName, "theNameFlag")
				fs.Set(FlagCommissionRate, "0.54")
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                ip,
				From:              "theNameFlag",
				Moniker:           "theNameFlag",
				ChainID:           chainID,
				NodeID:            nodeID,
				TrustNode:         true,
				PubKey:            sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Amount:            "100000000stake",
				CommissionRate:    "0.54",
				CommissionMaxRate: "0.2",
			},
		},
		{
			name: "Custom commission max rate",
			config: func() *cfg.Config {
				config := &cfg.Config{BaseConfig: cfg.TestBaseConfig()}
				config.Moniker = ""

				return config
			},
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(flags.FlagName, "theNameFlag")
				fs.Set(FlagCommissionMaxRate, "0.89")
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                ip,
				From:              "theNameFlag",
				Moniker:           "theNameFlag",
				ChainID:           chainID,
				NodeID:            nodeID,
				TrustNode:         true,
				PubKey:            sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Amount:            "100000000stake",
				CommissionRate:    "0.1",
				CommissionMaxRate: "0.89",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fs, _ := CreateValidatorMsgFlagSet(ip)
			fs.String(flags.FlagName, "", "name of private key with which to sign the gentx")

			tc.fsModify(fs)

			config := tc.config()

			cvCfg, err := PrepareConfigForTxCreateValidator(config, fs, nodeID, chainID, valPubKey)
			require.NoError(t, err)

			require.Equal(t, tc.expectedCfg, cvCfg)
		})
	}

}

func TestPrepareFlagsForTxCreateValidator(t *testing.T) {
	t.SkipNow()
	config, err := tcmd.ParseConfig()
	require.Nil(t, err)
	logger := log.NewNopLogger()
	ctx := server.NewContext(viper.New(), config, logger)

	valPubKey, _ := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, "cosmosvalconspub1zcjduepq7jsrkl9fgqk0wj3ahmfr8pgxj6vakj2wzn656s8pehh0zhv2w5as5gd80a")

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
		tt := tt
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) { runTest(t, tt, params) })
	}
}
