package v040_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
	v036supply "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_36"
	v038bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_38"
	v040bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_40"

	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	v040Codec := codec.New()
	cryptocodec.RegisterCrypto(v040Codec)
	v039auth.RegisterCodec(v040Codec)

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

	supply := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))

	bankGenState := v038bank.GenesisState{
		SendEnabled: true,
	}
	authGenState := v039auth.GenesisState{
		Accounts: v038auth.GenesisAccounts{acc1, vaac},
	}
	supplyGenState := v036supply.GenesisState{
		Supply: supply,
	}

	migrated := v040bank.Migrate(bankGenState, authGenState, supplyGenState)
	expected := `{
  "params": {
    "default_send_enabled": true
  },
  "balances": [
    {
      "address": "cosmos1xxkueklal9vejv9unqu80w9vptyepfa95pd53u",
      "coins": [
        {
          "denom": "stake",
          "amount": "50"
        }
      ]
    },
    {
      "address": "cosmos15v50ymp6n5dn73erkqtmq0u8adpl8d3ujv2e74",
      "coins": [
        {
          "denom": "stake",
          "amount": "50"
        }
      ]
    }
  ],
  "supply": [
    {
      "denom": "stake",
      "amount": "1000"
    }
  ],
  "denom_metadata": []
}`

	bz, err := v040Codec.MarshalJSONIndent(migrated, "", "  ")
	require.NoError(t, err)
	require.Equal(t, expected, string(bz))
}
