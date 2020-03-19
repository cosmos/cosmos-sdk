package keys

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBech32KeysOutput(t *testing.T) {
	tmpKey := secp256k1.GenPrivKey().PubKey()
	bechTmpKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, tmpKey)
	tmpAddr := sdk.AccAddress(tmpKey.Address().Bytes())

	multisigPks := multisig.NewPubKeyMultisigThreshold(1, []crypto.PubKey{tmpKey})
	multiInfo := NewMultiInfo("multisig", multisigPks)
	accAddr := sdk.AccAddress(multiInfo.GetPubKey().Address().Bytes())
	bechPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, multiInfo.GetPubKey())

	expectedOutput := NewKeyOutput(multiInfo.GetName(), multiInfo.GetType().String(), accAddr.String(), bechPubKey)
	expectedOutput.Threshold = 1
	expectedOutput.PubKeys = []multisigPubKeyOutput{{tmpAddr.String(), bechTmpKey, 1}}

	outputs, err := Bech32KeysOutput([]Info{multiInfo})
	require.NoError(t, err)
	require.Equal(t, expectedOutput, outputs[0])
}
