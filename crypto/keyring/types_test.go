package keyring_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_writeReadLedgerInfo(t *testing.T) {
	tmpKey := make([]byte, secp256k1.PubKeySize)
	hexPK := "035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A"
	bz, err := hex.DecodeString(hexPK)
	require.NoError(t, err)
	copy(tmpKey[:], bz)

	pk := &secp256k1.PubKey{Key: tmpKey}
	apk, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)
	path := hd.NewFundraiserParams(5, sdk.CoinType, 1)
	ledgerRecord := keyring.NewLedgerRecord(path)
	ledgerRecordItem := keyring.NewLedgerRecordItem(ledgerRecord)
	k := keyring.NewRecord("some_name", apk, ledgerRecordItem)
	//require.Equal(t, keyring.TypeLedger, kr.GetType())

	path = k.GetLedger().GetPath()
	require.Equal(t, "purpose:44 coinType:118 account:5 addressIndex:1 ", path.String())
	pubKey, err := k.GetPubKey()
	require.NoError(t, err)
	require.Equal(t,
		fmt.Sprintf("PubKeySecp256k1{%s}", hexPK),
		pubKey.String())

	// Serialize and restore
	encCfg := simapp.MakeTestEncodingConfig()
	serialized, err := encCfg.Marshaler.Marshal(k)
	require.NoError(t, err)
	var restoredRecord keyring.Record
	err = encCfg.Marshaler.Unmarshal(serialized, &restoredRecord)
	require.NoError(t, err)
	require.NotNil(t, restoredRecord)

	// Check both keys match
	require.Equal(t, k.GetName(), restoredRecord.GetName())
	require.Equal(t, k.GetType(), restoredRecord.GetType())
	//TODO fix error
	//restoredPubKey, err := restoredRecord.GetPubKey()
	//require.NoError(t, err)
	//require.Equal(t, pubKey, restoredPubKey)

	restoredPath := restoredRecord.GetLedger().GetPath()
	require.NoError(t, err)
	require.Equal(t, path, restoredPath)
}
