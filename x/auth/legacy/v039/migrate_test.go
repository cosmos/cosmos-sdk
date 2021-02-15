package v039_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v038"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
)

func TestMigrate(t *testing.T) {
	aminoCdc := codec.NewLegacyAmino()
	v039auth.RegisterLegacyAminoCodec(aminoCdc)

	pub1 := ed25519.GenPrivKeyFromSecret([]byte("acc1")).PubKey()
	pub2 := secp256k1.GenPrivKeyFromSecret([]byte("acc2")).PubKey()

	acc1 := v038auth.BaseAccount{
		Address:       sdk.AccAddress(pub1.Address()),
		Coins:         sdk.NewCoins(sdk.NewInt64Coin("stake", 400000)),
		Sequence:      1,
		AccountNumber: 1,
		PubKey:        pub1,
	}
	acc2 := v038auth.BaseAccount{
		Address:       sdk.AccAddress(pub2.Address()),
		Coins:         sdk.NewCoins(sdk.NewInt64Coin("stake", 400000)),
		Sequence:      2,
		AccountNumber: 2,
		PubKey:        pub2,
	}

	migrated := v039auth.Migrate(
		v038auth.GenesisState{
			Accounts: v038auth.GenesisAccounts{&acc1, &acc2},
		},
	)

	expectedAcc1 := v039auth.NewBaseAccount(acc1.Address, acc1.Coins, acc1.PubKey, acc1.AccountNumber, acc1.Sequence)
	expectedAcc2 := v039auth.NewBaseAccount(acc2.Address, acc2.Coins, acc2.PubKey, acc2.AccountNumber, acc2.Sequence)

	require.Equal(
		t, migrated, v039auth.GenesisState{
			Accounts: v038auth.GenesisAccounts{expectedAcc1, expectedAcc2},
		},
	)

	json, err := aminoCdc.MarshalJSONIndent(migrated, "", "  ")
	require.NoError(t, err)

	expectedJSON := `{
  "params": {
    "max_memo_characters": "0",
    "tx_sig_limit": "0",
    "tx_size_cost_per_byte": "0",
    "sig_verify_cost_ed25519": "0",
    "sig_verify_cost_secp256k1": "0"
  },
  "accounts": [
    {
      "type": "cosmos-sdk/Account",
      "value": {
        "address": "cosmos1j7skdhh9raxdmfhmcy2gxz8hgn0jnhfmujjsfe",
        "coins": [
          {
            "denom": "stake",
            "amount": "400000"
          }
        ],
        "public_key": {
          "type": "tendermint/PubKeyEd25519",
          "value": "eB0AcLMLKFRNFfh4XAAMstexfAIUQQCDnfjLZ2KJg+A="
        },
        "account_number": "1",
        "sequence": "1"
      }
    },
    {
      "type": "cosmos-sdk/Account",
      "value": {
        "address": "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
        "coins": [
          {
            "denom": "stake",
            "amount": "400000"
          }
        ],
        "public_key": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "AruDygh5HprMOpHOEato85dLgAsybMJVyxBGUa3KuWCr"
        },
        "account_number": "2",
        "sequence": "2"
      }
    }
  ]
}`
	require.Equal(t, expectedJSON, string(json))
}
