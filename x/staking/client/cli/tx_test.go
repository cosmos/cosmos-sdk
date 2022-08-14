package cli

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
)

func TestPrepareConfigForTxCreateValidator(t *testing.T) {
	chainID := "chainID"
	ip := "1.1.1.1"
	nodeID := "nodeID"
	privKey := ed25519.GenPrivKey()
	valPubKey := privKey.PubKey()
	moniker := "DefaultMoniker"
	mkTxValCfg := func(amount, commission, commissionMax, commissionMaxChange, minSelfDelegation string) TxCreateValidatorConfig {
		return TxCreateValidatorConfig{
			IP:                      ip,
			ChainID:                 chainID,
			NodeID:                  nodeID,
			PubKey:                  valPubKey,
			Moniker:                 moniker,
			Amount:                  amount,
			CommissionRate:          commission,
			CommissionMaxRate:       commissionMax,
			CommissionMaxChangeRate: commissionMaxChange,
			MinSelfDelegation:       minSelfDelegation,
		}
	}

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
			expectedCfg: mkTxValCfg(defaultAmount, "0.1", "0.2", "0.01", "1"),
		}, {
			name: "Custom amount",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagAmount, "2000stake")
			},
			expectedCfg: mkTxValCfg("2000stake", "0.1", "0.2", "0.01", "1"),
		}, {
			name: "Custom commission rate",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagCommissionRate, "0.54")
			},
			expectedCfg: mkTxValCfg(defaultAmount, "0.54", "0.2", "0.01", "1"),
		}, {
			name: "Custom commission max rate",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagCommissionMaxRate, "0.89")
			},
			expectedCfg: mkTxValCfg(defaultAmount, "0.1", "0.89", "0.01", "1"),
		}, {
			name: "Custom commission max change rate",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagCommissionMaxChangeRate, "0.55")
			},
			expectedCfg: mkTxValCfg(defaultAmount, "0.1", "0.2", "0.55", "1"),
		},
		{
			name: "Custom min self delegations",
			fsModify: func(fs *pflag.FlagSet) {
				fs.Set(FlagMinSelfDelegation, "0.33")
			},
			expectedCfg: mkTxValCfg(defaultAmount, "0.1", "0.2", "0.01", "0.33"),
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
