package keyring

import (
	"testing"

	"github.com/stretchr/testify/require"

	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/internal/protocdc"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBech32KeysOutput(t *testing.T) {
	tmpKey := secp256k1.GenPrivKey().PubKey()
	tmpAddr := sdk.AccAddress(tmpKey.Address().Bytes())
	multisigPk := kmultisig.NewLegacyAminoPubKey(1, []types.PubKey{tmpKey})

	multiInfo := NewMultiInfo("multisig", multisigPk)
	accAddr := sdk.AccAddress(multiInfo.GetPubKey().Address().Bytes())

	expectedOutput, err := NewKeyOutput(multiInfo.GetName(), multiInfo.GetType(), accAddr, multisigPk)
	require.NoError(t, err)
	expectedOutput.Threshold = 1
	tmpKeyBz, err := protocdc.MarshalJSON(tmpKey, nil)
	expectedOutput.PubKeys = []multisigPubKeyOutput{{tmpAddr.String(), string(tmpKeyBz), 1}}

	outputs, err := Bech32KeysOutput([]Info{multiInfo})
	require.NoError(t, err)
	require.Equal(t, expectedOutput.PubKeys, outputs[0].PubKeys)

	// expectedOutput is:
	// {"name":"multisig","type":"multi","address":"cosmos1dutd6e06gv4tfdypzyvdfuwct8cdtctsjr0nl3","pubkey":"{\"threshold\":1,\"public_keys\":[{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A5edrlIT2MWsp4SYh+KvGfjxmKw7NNxga7K23faEEmL/\"}]}","threshold":1,"pubkeys":[{"address":"cosmos13533lkpyxrgttefm34f35c89v2yrv3ugkns5p6","pubkey":"{\"key\":\"A5edrlIT2MWsp4SYh+KvGfjxmKw7NNxga7K23faEEmL/\"}","weight":1}]}
}
