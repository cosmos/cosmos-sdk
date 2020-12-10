package v040_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034 "github.com/cosmos/cosmos-sdk/x/auth/legacy/v034"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v038"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONMarshaler(encodingConfig.Marshaler)

	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	// BaseAccount
	pk1 := secp256k1.GenPrivKeyFromSecret([]byte("acc1")).PubKey()
	acc1 := v039auth.NewBaseAccount(sdk.AccAddress(pk1.Address()), coins, pk1, 1, 0)

	// ModuleAccount
	pk2 := secp256k1.GenPrivKeyFromSecret([]byte("acc2")).PubKey()
	acc2 := v039auth.NewModuleAccount(
		v039auth.NewBaseAccount(sdk.AccAddress(pk2.Address()), coins, pk2, 1, 0),
		"module2",
		"permission2",
	)

	// BaseVestingAccount
	pk3 := secp256k1.GenPrivKeyFromSecret([]byte("acc3")).PubKey()
	acc3 := v039auth.NewBaseVestingAccount(
		v039auth.NewBaseAccount(sdk.AccAddress(pk3.Address()), coins, pk3, 1, 0),
		coins, coins, coins,
		1580309973,
	)

	// ContinuousVestingAccount
	pk4 := secp256k1.GenPrivKeyFromSecret([]byte("acc4")).PubKey()
	acc4 := v039auth.NewContinuousVestingAccountRaw(
		v039auth.NewBaseVestingAccount(v039auth.NewBaseAccount(sdk.AccAddress(pk4.Address()), coins, pk4, 1, 0), coins, nil, nil, 3160620846),
		1580309974,
	)

	// PeriodicVestingAccount
	pk5 := secp256k1.GenPrivKeyFromSecret([]byte("acc5")).PubKey()
	acc5 := &v039auth.PeriodicVestingAccount{
		BaseVestingAccount: v039auth.NewBaseVestingAccount(v039auth.NewBaseAccount(sdk.AccAddress(pk5.Address()), coins, pk5, 1, 0), coins, nil, nil, 3160620846),
		StartTime:          1580309975,
		VestingPeriods:     v039auth.Periods{v039auth.Period{Length: 32, Amount: coins}},
	}

	// DelayedVestingAccount
	pk6 := secp256k1.GenPrivKeyFromSecret([]byte("acc6")).PubKey()
	acc6 := &v039auth.DelayedVestingAccount{
		BaseVestingAccount: v039auth.NewBaseVestingAccount(v039auth.NewBaseAccount(sdk.AccAddress(pk6.Address()), coins, pk6, 1, 0), coins, nil, nil, 3160620846),
	}

	// BaseAccount with nil pubkey (coming from older genesis).
	pk7 := secp256k1.GenPrivKeyFromSecret([]byte("acc7")).PubKey()
	acc7 := v039auth.NewBaseAccount(sdk.AccAddress(pk7.Address()), coins, nil, 1, 0)

	gs := v039auth.GenesisState{
		Params: v034.Params{
			MaxMemoCharacters:      10,
			TxSigLimit:             20,
			TxSizeCostPerByte:      30,
			SigVerifyCostED25519:   40,
			SigVerifyCostSecp256k1: 50,
		},
		Accounts: v038auth.GenesisAccounts{acc1, acc2, acc3, acc4, acc5, acc6, acc7},
	}

	migrated := v040auth.Migrate(gs)
	expected := `{
  "accounts": [
    {
      "@type": "/cosmos.auth.v1beta1.BaseAccount",
      "account_number": "1",
      "address": "cosmos13syh7de9xndv9wmklccpfvc0d8dcyvay4s6z6l",
      "pub_key": {
        "@type": "/cosmos.crypto.secp256k1.PubKey",
        "key": "A8oWyJkohwy8XZ0Df92jFMBTtTPMvYJplYIrlEHTKPYk"
      },
      "sequence": "0"
    },
    {
      "@type": "/cosmos.auth.v1beta1.ModuleAccount",
      "base_account": {
        "account_number": "1",
        "address": "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
        "pub_key": {
          "@type": "/cosmos.crypto.secp256k1.PubKey",
          "key": "AruDygh5HprMOpHOEato85dLgAsybMJVyxBGUa3KuWCr"
        },
        "sequence": "0"
      },
      "name": "module2",
      "permissions": [
        "permission2"
      ]
    },
    {
      "@type": "/cosmos.vesting.v1beta1.BaseVestingAccount",
      "base_account": {
        "account_number": "1",
        "address": "cosmos18hnp9fjflrkeeqn4gmhjhzljusxzmjeartdckw",
        "pub_key": {
          "@type": "/cosmos.crypto.secp256k1.PubKey",
          "key": "A5aEFDIdQHh0OYmNXNv1sHBNURDWWgVkXC2IALcWLLwJ"
        },
        "sequence": "0"
      },
      "delegated_free": [
        {
          "amount": "50",
          "denom": "stake"
        }
      ],
      "delegated_vesting": [
        {
          "amount": "50",
          "denom": "stake"
        }
      ],
      "end_time": "1580309973",
      "original_vesting": [
        {
          "amount": "50",
          "denom": "stake"
        }
      ]
    },
    {
      "@type": "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
      "base_vesting_account": {
        "base_account": {
          "account_number": "1",
          "address": "cosmos1t9kvvejvk6hjtddx6antck39s206csqduq3ke3",
          "pub_key": {
            "@type": "/cosmos.crypto.secp256k1.PubKey",
            "key": "AoXDzxwTnljemHxfnJcwrKqODBP6Q2l3K3U3UhVDzyah"
          },
          "sequence": "0"
        },
        "delegated_free": [],
        "delegated_vesting": [],
        "end_time": "3160620846",
        "original_vesting": [
          {
            "amount": "50",
            "denom": "stake"
          }
        ]
      },
      "start_time": "1580309974"
    },
    {
      "@type": "/cosmos.vesting.v1beta1.PeriodicVestingAccount",
      "base_vesting_account": {
        "base_account": {
          "account_number": "1",
          "address": "cosmos1s4ss9zquz7skvguechzlk3na635jdrecl0sgy2",
          "pub_key": {
            "@type": "/cosmos.crypto.secp256k1.PubKey",
            "key": "A2a4P4TQ1OKzpfu0eKnCoEtmTvoiclSx0G9higenUGws"
          },
          "sequence": "0"
        },
        "delegated_free": [],
        "delegated_vesting": [],
        "end_time": "3160620846",
        "original_vesting": [
          {
            "amount": "50",
            "denom": "stake"
          }
        ]
      },
      "start_time": "1580309975",
      "vesting_periods": [
        {
          "amount": [
            {
              "amount": "50",
              "denom": "stake"
            }
          ],
          "length": "32"
        }
      ]
    },
    {
      "@type": "/cosmos.vesting.v1beta1.DelayedVestingAccount",
      "base_vesting_account": {
        "base_account": {
          "account_number": "1",
          "address": "cosmos1mcc6rwrj4hswf8p9ct82c7lmf77w9tuk07rha4",
          "pub_key": {
            "@type": "/cosmos.crypto.secp256k1.PubKey",
            "key": "A4tuAfmZlhjK5cjp6ImR704miybHnITVNOyJORdDPFu3"
          },
          "sequence": "0"
        },
        "delegated_free": [],
        "delegated_vesting": [],
        "end_time": "3160620846",
        "original_vesting": [
          {
            "amount": "50",
            "denom": "stake"
          }
        ]
      }
    },
    {
      "@type": "/cosmos.auth.v1beta1.BaseAccount",
      "account_number": "1",
      "address": "cosmos16ydaqh0fcnh4qt7a3jme4mmztm2qel5axcpw00",
      "pub_key": null,
      "sequence": "0"
    }
  ],
  "params": {
    "max_memo_characters": "10",
    "sig_verify_cost_ed25519": "40",
    "sig_verify_cost_secp256k1": "50",
    "tx_sig_limit": "20",
    "tx_size_cost_per_byte": "30"
  }
}`

	bz, err := clientCtx.JSONMarshaler.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "  ")
	require.NoError(t, err)

	require.Equal(t, expected, string(indentedBz))
}
