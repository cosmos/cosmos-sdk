package rosetta

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"testing"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/require"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func TestLaunchpad_ConstructionCombine(t *testing.T) {
	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
		AddrPrefix: "test",
	}

	adapter := newAdapter(cosmos.NewClient(""), tendermint.NewClient(""), properties)
	bz, err := ioutil.ReadFile("./testdata/unsigned-tx.json")
	require.NoError(t, err)

	var stdTx auth.StdTx
	err = Codec.UnmarshalJSON(bz, &stdTx)
	require.NoError(t, err)
	txBytes, err := Codec.MarshalJSON(stdTx)
	require.NoError(t, err)
	txHex := hex.EncodeToString(txBytes)

	pk, err := secp256k1.NewPrivateKey(secp256k1.S256())
	require.NoError(t, err)
	var combineRes, combineErr = adapter.ConstructionCombine(context.Background(), &types.ConstructionCombineRequest{
		UnsignedTransaction: txHex,
		Signatures: []*types.Signature{{
			SigningPayload: &types.SigningPayload{
				Address: "test1qrv8g4hwt4z6ds8mednhhgx907wug9d6y8n9jy",
				Bytes:   txBytes,
			},
			PublicKey: &types.PublicKey{
				CurveType: types.Secp256k1,
				Bytes:     pk.PubKey().SerializeCompressed(),
			},
			SignatureType: types.Ecdsa,
			// uses random bytes as signing is out of scope for rosetta
			Bytes: txBytes,
		},
		}})
	require.Nil(t, combineErr)
	require.NotNil(t, combineRes)

	bz, err = hex.DecodeString(combineRes.SignedTransaction)
	require.NoError(t, err)
	var signedStdTx auth.StdTx
	err = Codec.UnmarshalJSON(bz, &signedStdTx)
	require.NoError(t, err)
	require.Equal(t, stdTx.GetSigners(), signedStdTx.GetSigners())
}
