package keyring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBech32KeysOutput(t *testing.T) {
	tmpKey := secp256k1.GenPrivKey().PubKey()
	tmpAddr := sdk.AccAddress(tmpKey.Address().Bytes())

	multisigPks := kmultisig.NewLegacyAminoPubKey(1, []types.PubKey{tmpKey})
	multiInfo := NewMultiInfo("multisig", multisigPks)
	accAddr := sdk.AccAddress(multiInfo.GetPubKey().Address().Bytes())

	fmt.Println(multiInfo)
	// TODO:
	// require.True(t, accAddr.Equals(tmpAddr), "- %s\n+ %s", tmpAddr, accAddr)

	expectedOutput, err := NewKeyOutput(multiInfo.GetName(), multiInfo.GetType(), accAddr, tmpKey)
	require.NoError(t, err)
	expectedOutput.Threshold = 1

	// TODO: check: pkStr = tmpKey.String()
	expectedOutput.PubKeys = []multisigPubKeyOutput{{tmpAddr.String(), expectedOutput.PubKey, 1}}

	outputs, err := Bech32KeysOutput([]Info{multiInfo})
	require.NoError(t, err)
	require.Equal(t, expectedOutput, outputs[0])
}
