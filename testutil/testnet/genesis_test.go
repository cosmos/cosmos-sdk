package testnet_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/testnet"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/stretchr/testify/require"
)

func TestGenesisBuilder_GenesisTime(t *testing.T) {
	t.Run("defaults to current time", func(t *testing.T) {
		before := time.Now()
		time.Sleep(time.Millisecond) // So that the genesis time will be strictly after "before".
		gb := testnet.NewGenesisBuilder()
		time.Sleep(time.Millisecond) // So that the genesis time will be strictly before "after".
		after := time.Now()

		var gts string
		require.NoError(t, json.Unmarshal(gb.JSON()["genesis_time"], &gts))

		gt, err := time.Parse(time.RFC3339Nano, gts)
		require.NoError(t, err)

		require.True(t, gt.After(before))
		require.True(t, after.After(gt))
	})

	t.Run("can be set to arbitrary time", func(t *testing.T) {
		want := time.Date(2023, 3, 27, 9, 50, 23, 0, time.UTC)

		gb := testnet.NewGenesisBuilder().GenesisTime(want)

		var gts string
		require.NoError(t, json.Unmarshal(gb.JSON()["genesis_time"], &gts))

		gt, err := time.Parse(time.RFC3339Nano, gts)
		require.NoError(t, err)

		require.True(t, gt.Equal(want))
	})
}

func TestGenesisBuilder_InitialHeight(t *testing.T) {
	t.Run("defaults to 1", func(t *testing.T) {
		var ih string
		require.NoError(
			t,
			json.Unmarshal(
				testnet.NewGenesisBuilder().JSON()["initial_height"],
				&ih,
			),
		)

		require.Equal(t, ih, "1")
	})

	t.Run("can be set to arbitrary value", func(t *testing.T) {
		var ih string
		require.NoError(
			t,
			json.Unmarshal(
				testnet.NewGenesisBuilder().InitialHeight(12345).JSON()["initial_height"],
				&ih,
			),
		)

		require.Equal(t, ih, "12345")
	})
}

func TestGenesisBuilder_ChainID(t *testing.T) {
	// No default.
	gb := testnet.NewGenesisBuilder()
	m := gb.JSON()
	_, ok := m["chain_id"]
	require.False(t, ok)

	var id string
	require.NoError(
		t,
		json.Unmarshal(
			gb.ChainID("my-chain").JSON()["chain_id"],
			&id,
		),
	)
	require.Equal(t, id, "my-chain")
}

// Use known keys and addresses to assert that correct validator and delegator keys
// occur in the expected locations (i.e. we didn't mistakenly swap the keys anywhere).
func TestGenesisBuilder_GentxAddresses(t *testing.T) {
	const valSecret0 = "val-secret-0"
	const valAddr0 = "3F3B076353767F046477A6E0982F808C24D1870A"
	const valPubKey0 = "ZhVhrOUHnUwYw/GlBSBrw/0X6A261gchCRYkAxGF2jk="
	valKey0 := cmted25519.GenPrivKeyFromSecret([]byte(valSecret0))
	if addr := valKey0.PubKey().Address().String(); addr != valAddr0 {
		t.Fatalf("unexpected address %q for validator key 0 (expected %q)", addr, valAddr0)
	}
	if pub := base64.StdEncoding.EncodeToString(valKey0.PubKey().Bytes()); pub != valPubKey0 {
		t.Fatalf("unexpected public key %q for validator key 0 (expected %q)", pub, valAddr0)
	}

	const delSecret0 = "del-secret-0"
	const delAddr0 = "30D7E04DA313C31B59A46408494B4272F0A9A256"
	const delPubKey0 = "Aol+ZF9xBuZmYJrT1QFLpZBvSfr/zEKifWyg0Xi1tsFV"
	const delAccAddr0 = "cosmos1xrt7qndrz0p3kkdyvsyyjj6zwtc2ngjky8dcpe"
	delKey0 := secp256k1.GenPrivKeyFromSecret([]byte(delSecret0))
	if addr := delKey0.PubKey().Address().String(); addr != delAddr0 {
		t.Fatalf("unexpected address %q for delegator key 0 (expected %q)", addr, delAddr0)
	}
	if pub := base64.StdEncoding.EncodeToString(delKey0.PubKey().Bytes()); pub != delPubKey0 {
		t.Fatalf("unexpected public key %q for delegator key 0 (expected %q)", pub, delAddr0)
	}
	da, err := bech32.ConvertAndEncode("cosmos", delKey0.PubKey().Address().Bytes())
	require.NoError(t, err)
	if da != delAccAddr0 {
		t.Fatalf("unexpected account address %q for delegator key 0 (expected %q)", da, delAccAddr0)
	}

	valPKs := testnet.ValidatorPrivKeys{
		&testnet.ValidatorPrivKey{
			Val: valKey0,
			Del: delKey0,
		},
	}
	cmtVals := valPKs.CometGenesisValidators()
	stakingVals := cmtVals.StakingValidators()
	valBaseAccounts := stakingVals.BaseAccounts()

	b := testnet.NewGenesisBuilder().
		ChainID("my-test-chain").
		DefaultAuthParams().
		Consensus(nil, cmtVals).
		BaseAccounts(valBaseAccounts, nil).
		StakingWithDefaultParams(stakingVals, nil)

	for i, v := range valPKs {
		b.GenTx(*v.Del, cmtVals[i].V, sdk.NewCoin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction))
	}

	var g struct {
		Consensus struct {
			Validators []struct {
				Address string `json:"address"`
				PubKey  struct {
					Value string `json:"value"`
				} `json:"pub_key"`
			} `json:"validators"`
		} `json:"consensus"`

		AppState struct {
			Genutil struct {
				GenTxs []struct {
					Body struct {
						Messages []struct {
							Type             string `json:"@type"`
							DelegatorAddress string `json:"delegator_address"`
							ValidatorAddress string `json:"validator_address"`
							PubKey           struct {
								Key string `json:"key"`
							} `json:"pubkey"`
						} `json:"messages"`
					} `json:"body"`
					AuthInfo struct {
						SignerInfos []struct {
							PublicKey struct {
								Key string `json:"key"`
							} `json:"public_key"`
						} `json:"signer_infos"`
					} `json:"auth_info"`
				} `json:"gen_txs"`
			} `json:"genutil"`

			Auth struct {
				Accounts []struct {
					Address string `json:"address"`
					PubKey  struct {
						Key string `json:"key"`
					} `json:"pub_key"`
				} `json:"accounts"`
			} `json:"auth"`
		} `json:"app_state"`
	}
	if err := json.Unmarshal(b.Encode(), &g); err != nil {
		t.Fatal(err)
	}

	// Validator encoded as expected.
	vals := g.Consensus.Validators
	require.Equal(t, vals[0].Address, valAddr0)
	require.Equal(t, vals[0].PubKey.Value, valPubKey0)

	// Public keys on gentx message match correct keys (no ed25519/secp256k1 mismatch).
	gentxs := g.AppState.Genutil.GenTxs
	require.Equal(t, gentxs[0].Body.Messages[0].PubKey.Key, valPubKey0)
	require.Equal(t, gentxs[0].AuthInfo.SignerInfos[0].PublicKey.Key, delPubKey0)

	// The only base account in this genesis, matches the secp256k1 key.
	acct := g.AppState.Auth.Accounts[0]
	require.Equal(t, acct.Address, delAccAddr0)
	require.Equal(t, acct.PubKey.Key, delPubKey0)
}

func ExampleGenesisBuilder() {
	const nVals = 4

	// Generate random private keys for validators.
	valPKs := testnet.NewValidatorPrivKeys(nVals)

	// Produce the comet representation of those validators.
	cmtVals := valPKs.CometGenesisValidators()

	stakingVals := cmtVals.StakingValidators()

	// Configure a new genesis builder
	// with a fairly thorough set of defaults.
	//
	// If you only ever need defaults, you can use DefaultGenesisBuilderOnlyValidators().
	b := testnet.NewGenesisBuilder().
		ChainID("my-chain-id").
		DefaultAuthParams().
		Consensus(nil, cmtVals).
		BaseAccounts(stakingVals.BaseAccounts(), nil).
		DefaultStaking().
		BankingWithDefaultParams(stakingVals.Balances(), nil, nil, nil).
		DefaultDistribution().
		DefaultMint().
		DefaultSlashing()

	for i := range stakingVals {
		b.GenTx(*valPKs[i].Del, cmtVals[i].V, sdk.NewCoin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction))
	}

	// Now, you can access b.JSON() if you need to make further modifications
	// not (yet) supported by the GenesisBuilder API,
	// or you can use b.Encode() for the serialzed JSON of the genesis.
}
