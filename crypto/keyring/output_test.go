package keyring

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"
)

func TestBech32KeysOutput(t *testing.T) {
	tmpKey := secp256k1.GenPrivKey().PubKey()
	bechTmpKey := legacybech32.MustBech32ifyPubKey(legacybech32.Bech32PubKeyTypeAccPub, tmpKey)
	tmpAddr := sdk.AccAddress(tmpKey.Address().Bytes())

	multisigPks := kmultisig.NewLegacyAminoPubKey(1, []crypto.PubKey{tmpKey})
	multiInfo := NewMultiInfo("multisig", multisigPks)
	accAddr := sdk.AccAddress(multiInfo.GetPubKey().Address().Bytes())
	bechPubKey := legacybech32.MustBech32ifyPubKey(legacybech32.Bech32PubKeyTypeAccPub, multiInfo.GetPubKey())

	expectedOutput := NewKeyOutput(multiInfo.GetName(), multiInfo.GetType().String(), accAddr.String(), bechPubKey)
	expectedOutput.Threshold = 1
	expectedOutput.PubKeys = []multisigPubKeyOutput{{tmpAddr.String(), bechTmpKey, 1}}

	outputs, err := Bech32KeysOutput([]Info{multiInfo})
	require.NoError(t, err)
	require.Equal(t, expectedOutput, outputs[0])
}
