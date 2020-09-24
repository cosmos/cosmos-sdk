package v040_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034 "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_34"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
	v040 "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_40"
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONMarshaler(encodingConfig.Marshaler)

	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))
	addr1, _ := sdk.AccAddressFromBech32("cosmos1xxkueklal9vejv9unqu80w9vptyepfa95pd53u")
	acc1 := v039auth.NewBaseAccount(addr1, coins, secp256k1.GenPrivKey().PubKey(), 1, 0)

	addr2, _ := sdk.AccAddressFromBech32("cosmos15v50ymp6n5dn73erkqtmq0u8adpl8d3ujv2e74")
	vaac := v039auth.NewContinuousVestingAccountRaw(
		v039auth.NewBaseVestingAccount(v039auth.NewBaseAccount(addr2, coins, secp256k1.GenPrivKey().PubKey(), 1, 0), coins, nil, nil, 3160620846),
		1580309972,
	)

	gs := v039auth.GenesisState{
		Params: v034.Params{
			MaxMemoCharacters:      10,
			TxSigLimit:             10,
			TxSizeCostPerByte:      10,
			SigVerifyCostED25519:   10,
			SigVerifyCostSecp256k1: 10,
		},
		Accounts: v038auth.GenesisAccounts{acc1, vaac},
	}

	migrated := v040.Migrate(gs)
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
      "type": "cosmos-sdk/Account",
      "value": {
        "address": "cosmos1xxkueklal9vejv9unqu80w9vptyepfa95pd53u",
        "public_key": null,
        "account_number": "1",
        "sequence": "0"
      }
    },
    {
      "type": "cosmos-sdk/ContinuousVestingAccount",
      "value": {
        "address": "cosmos15v50ymp6n5dn73erkqtmq0u8adpl8d3ujv2e74",
        "public_key": null,
        "account_number": "1",
        "sequence": "0",
        "original_vesting": [
          {
            "denom": "stake",
            "amount": "50"
          }
        ],
        "delegated_free": [],
        "delegated_vesting": [],
        "end_time": "3160620846",
        "start_time": "1580309972"
      }
    }
  ]
}`

	fmt.Printf("migrated=%+v\n", migrated.Accounts)
	bz, err := clientCtx.JSONMarshaler.MarshalJSON(migrated)
	require.NoError(t, err)
	require.Equal(t, expected, string(bz))
}
