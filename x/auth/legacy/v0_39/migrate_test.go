package v039_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_34"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v039 "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"

	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	v039Codec := codec.New()
	codec.RegisterCrypto(v039Codec)
	v038auth.RegisterCodec(v039Codec)

	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))
	addr1, _ := sdk.AccAddressFromBech32("cosmos1xxkueklal9vejv9unqu80w9vptyepfa95pd53u")
	acc1 := v038auth.NewBaseAccount(addr1, coins, nil, 1, 0)

	addr2, _ := sdk.AccAddressFromBech32("cosmos15v50ymp6n5dn73erkqtmq0u8adpl8d3ujv2e74")
	vaac := v038auth.NewContinuousVestingAccountRaw(
		v038auth.NewBaseVestingAccount(
			v038auth.NewBaseAccount(addr2, coins, nil, 1, 0), coins, nil, nil, 3160620846,
		),
		1580309972,
	)

	gs := v038auth.GenesisState{
		Params: v0_34.Params{
			MaxMemoCharacters:      10,
			TxSigLimit:             10,
			TxSizeCostPerByte:      10,
			SigVerifyCostED25519:   10,
			SigVerifyCostSecp256k1: 10,
		},
		Accounts: v038auth.GenesisAccounts{acc1, vaac},
	}

	migrated := v039.Migrate(gs)
	expected := `{
  "params": {
    "max_memo_characters": "10",
    "tx_sig_limit": "10",
    "tx_size_cost_per_byte": "10",
    "sig_verify_cost_ed25519": "10",
    "sig_verify_cost_secp256k1": "10"
  },
  "accounts": [
    {
      "type": "cosmos-sdk/BaseAccount",
      "value": {
        "address": "cosmos1xxkueklal9vejv9unqu80w9vptyepfa95pd53u",
        "public_key": "",
        "account_number": 1,
        "sequence": 0
      }
    },
    {
      "type": "cosmos-sdk/ContinuousVestingAccount",
      "value": {
        "address": "cosmos15v50ymp6n5dn73erkqtmq0u8adpl8d3ujv2e74",
        "public_key": "",
        "account_number": 1,
        "sequence": 0,
        "original_vesting": [
          {
            "denom": "stake",
            "amount": "50"
          }
        ],
        "delegated_free": [],
        "delegated_vesting": [],
        "end_time": 3160620846,
        "start_time": 1580309972
      }
    }
  ]
}`

	bz, err := v039Codec.MarshalJSONIndent(migrated, "", "  ")
	require.NoError(t, err)
	require.Equal(t, expected, string(bz))
}
