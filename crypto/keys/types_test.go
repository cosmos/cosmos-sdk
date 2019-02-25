package keys

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func Test_writeReadLedgerInfo(t *testing.T) {
	var tmpKey secp256k1.PubKeySecp256k1
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(tmpKey[:], bz)

	lInfo := ledgerInfo{
		"some_name",
		tmpKey,
		*hd.NewFundraiserParams(5, 1)}
	assert.Equal(t, TypeLedger, lInfo.GetType())
	assert.Equal(t, "44'/118'/5'/0/1", lInfo.GetPath().String())
	assert.Equal(t,
		"cosmospub1addwnpepqddddqg2glc8x4fl7vxjlnr7p5a3czm5kcdp4239sg6yqdc4rc2r5wmxv8p",
		types.MustBech32ifyAccPub(lInfo.GetPubKey()))

	// Serialize and restore
	serialized := writeInfo(lInfo)
	restoredInfo, err := readInfo(serialized)
	assert.NoError(t, err)
	assert.NotNil(t, restoredInfo)

	// Check both keys match
	assert.Equal(t, lInfo.GetName(), restoredInfo.GetName())
	assert.Equal(t, lInfo.GetType(), restoredInfo.GetType())
	assert.Equal(t, lInfo.GetPubKey(), restoredInfo.GetPubKey())

	assert.Equal(t, lInfo.GetPath(), restoredInfo.(ledgerInfo).GetPath())
}
