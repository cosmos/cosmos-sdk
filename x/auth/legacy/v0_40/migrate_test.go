package v040_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034 "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_34"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_40"
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONMarshaler(encodingConfig.Marshaler)

	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))
	pk1 := secp256k1.GenPrivKeyFromSecret([]byte("acc1")).PubKey()
	addr1 := sdk.AccAddress(pk1.Address())
	acc1 := v039auth.NewBaseAccount(addr1, coins, pk1, 1, 0)

	pk2 := secp256k1.GenPrivKeyFromSecret([]byte("acc2")).PubKey()
	addr2 := sdk.AccAddress(pk1.Address())
	acc2 := v039auth.NewContinuousVestingAccountRaw(
		v039auth.NewBaseVestingAccount(v039auth.NewBaseAccount(addr2, coins, pk2, 1, 0), coins, nil, nil, 3160620846),
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
		Accounts: v038auth.GenesisAccounts{acc1, acc2},
	}

	migrated := v040auth.Migrate(gs)
	expected := `{"params":{"max_memo_characters":"10","tx_sig_limit":"10","tx_size_cost_per_byte":"10","sig_verify_cost_ed25519":"10","sig_verify_cost_secp256k1":"10"},"accounts":[{"@type":"/cosmos.auth.v1beta1.BaseAccount","address":"cosmos13syh7de9xndv9wmklccpfvc0d8dcyvay4s6z6l","pub_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A8oWyJkohwy8XZ0Df92jFMBTtTPMvYJplYIrlEHTKPYk"},"account_number":"1","sequence":"0"},{"@type":"/cosmos.vesting.v1beta1.ContinuousVestingAccount","base_vesting_account":{"base_account":{"address":"cosmos13syh7de9xndv9wmklccpfvc0d8dcyvay4s6z6l","pub_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AruDygh5HprMOpHOEato85dLgAsybMJVyxBGUa3KuWCr"},"account_number":"1","sequence":"0"},"original_vesting":[{"denom":"stake","amount":"50"}],"delegated_free":[],"delegated_vesting":[],"end_time":"3160620846"},"start_time":"3160620846"}]}`

	bz, err := clientCtx.JSONMarshaler.MarshalJSON(migrated)
	require.NoError(t, err)
	require.Equal(t, expected, string(bz))
}
