package v040_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v038bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_38"
	v039bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_39"

	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	v039Codec := codec.New()
	cryptocodec.RegisterCrypto(v039Codec)
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

	bankGenState := v038bank.GenesisState{
		SendEnabled: true,
	}
	authGenState := v038auth.GenesisState{
		Accounts: v038auth.GenesisAccounts{acc1, vaac},
	}

	migrated := v039bank.Migrate(bankGenState, authGenState)
	expected := `{
  "send_enabled": true,
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
  ]
}`

	bz, err := v039Codec.MarshalJSONIndent(migrated, "", "  ")
	require.NoError(t, err)
	require.Equal(t, expected, string(bz))
}
