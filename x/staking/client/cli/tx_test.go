package cli

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPrepareConfigForTxCreateValidator(t *testing.T) {
	chainID := "chainID"
	ip := "1.1.1.1"
	nodeID := "nodeID"
	valPubKey, _ := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, "cosmosvalconspub1zcjduepq7jsrkl9fgqk0wj3ahmfr8pgxj6vakj2wzn656s8pehh0zhv2w5as5gd80a")
	moniker := "DefaultMoniker"

	tests := []struct {
		name        string
		fsModify    func(fs *pflag.FlagSet)
		expectedCfg TxCreateValidatorConfig
	}{
		{
			name: "all defaults",
			fsModify: func(fs *pflag.FlagSet) {
				return
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                      ip,
				ChainID:                 chainID,
				NodeID:                  nodeID,
				PubKey:                  sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Moniker:                 moniker,
				Amount:                  defaultAmount,
				CommissionRate:          "0.1",
				CommissionMaxRate:       "0.2",
				CommissionMaxChangeRate: "0.01",
				MinSelfDelegation:       "1",
			},
		},
		{
			name: "Custom amount",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagAmount, "2000stake")
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                      ip,
				Moniker:                 moniker,
				ChainID:                 chainID,
				NodeID:                  nodeID,
				PubKey:                  sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Amount:                  "2000stake",
				CommissionRate:          "0.1",
				CommissionMaxRate:       "0.2",
				CommissionMaxChangeRate: "0.01",
				MinSelfDelegation:       "1",
			},
		},
		{
			name: "Custom commission rate",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagCommissionRate, "0.54")
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                      ip,
				Moniker:                 moniker,
				ChainID:                 chainID,
				NodeID:                  nodeID,
				PubKey:                  sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Amount:                  defaultAmount,
				CommissionRate:          "0.54",
				CommissionMaxRate:       "0.2",
				CommissionMaxChangeRate: "0.01",
				MinSelfDelegation:       "1",
			},
		},
		{
			name: "Custom commission max rate",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagCommissionMaxRate, "0.89")
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                      ip,
				Moniker:                 moniker,
				ChainID:                 chainID,
				NodeID:                  nodeID,
				PubKey:                  sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Amount:                  defaultAmount,
				CommissionRate:          "0.1",
				CommissionMaxRate:       "0.89",
				CommissionMaxChangeRate: "0.01",
				MinSelfDelegation:       "1",
			},
		},
		{
			name: "Custom commission max change rate",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagCommissionMaxChangeRate, "0.55")
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                      ip,
				Moniker:                 moniker,
				ChainID:                 chainID,
				NodeID:                  nodeID,
				PubKey:                  sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Amount:                  defaultAmount,
				CommissionRate:          "0.1",
				CommissionMaxRate:       "0.2",
				CommissionMaxChangeRate: "0.55",
				MinSelfDelegation:       "1",
			},
		},
		{
			name: "Custom min self delegations",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagMinSelfDelegation, "0.33")
			},
			expectedCfg: TxCreateValidatorConfig{
				IP:                      ip,
				Moniker:                 moniker,
				ChainID:                 chainID,
				NodeID:                  nodeID,
				PubKey:                  sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey),
				Amount:                  defaultAmount,
				CommissionRate:          "0.1",
				CommissionMaxRate:       "0.2",
				CommissionMaxChangeRate: "0.01",
				MinSelfDelegation:       "0.33",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fs, _ := CreateValidatorMsgFlagSet(ip)
			fs.String(flags.FlagName, "", "name of private key with which to sign the gentx")

			tc.fsModify(fs)

			cvCfg, err := PrepareConfigForTxCreateValidator(fs, moniker, nodeID, chainID, valPubKey)
			require.NoError(t, err)

			require.Equal(t, tc.expectedCfg, cvCfg)
		})
	}
}
