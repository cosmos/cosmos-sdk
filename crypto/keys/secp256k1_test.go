package keys

import (
	"bytes"
	"crypto/subtle"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	underlyingSecp256k1 "github.com/btcsuite/btcd/btcec"
)

type keyData struct {
	priv string
	pub  string
	addr string
}

var secpDataTable = []keyData{
	{
		priv: "a96e62ed3955e65be32703f12d87b6b5cf26039ecfa948dc5107a495418e5330",
		pub:  "02950e1cdfcb133d6024109fd489f734eeb4502418e538c28481f22bce276f248c",
		addr: "1CKZ9Nx4zgds8tU7nJHotKSDr4a9bYJCa3",
	},
}

func TestPubKeySecp256k1Address(t *testing.T) {
	for _, d := range secpDataTable {
		privB, _ := hex.DecodeString(d.priv)
		pubB, _ := hex.DecodeString(d.pub)
		addrBbz, _, err := base58.CheckDecode(d.addr)
		require.NoError(t, err)
		addrB := crypto.Address(addrBbz)

		priv := PrivKeySecp256K1{bytes: privB}
		pubKey := priv.PubKey()
		pubT, ok := pubKey.(PubKeySecp256K1)
		require.True(t, ok)
		pub := pubT.bytes
		addr := pubKey.Address()

		require.Equal(t, pub, pubB, "Expected pub keys to match")
		require.Equal(t, addr, addrB, "Expected addresses to match")
	}
}

func TestSignAndValidateSecp256k1(t *testing.T) {
	privKey, err := GenPrivKey(SECP256K1)
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)

	require.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	sig[3] ^= byte(0x01)

	require.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestSecp256k1Compatibility(t *testing.T) {
	tmPrivKey := secp256k1.GenPrivKey()
	tmPubKey := tmPrivKey.PubKey()

	privKey, err := GenPrivKey(SECP256K1)
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	require.Equal(t, sdk.AccAddress(tmPubKey.Address()).String(), sdk.AccAddress(pubKey.Address()).String())
	require.True(t, bytes.Equal(tmPubKey.Bytes()[:], pubKey.Bytes()))
	require.True(t, subtle.ConstantTimeCompare(privKey.Bytes()[:], privKey.Bytes()) == 1)
}

func TestGenPrivKeySecp256k1(t *testing.T) {
	// curve oder N
	N := underlyingSecp256k1.S256().N
	tests := []struct {
		name   string
		secret []byte
	}{
		{"empty secret", []byte{}},
		{
			"some long secret",
			[]byte("We live in a society exquisitely dependent on science and technology, " +
				"in which hardly anyone knows anything about science and technology."),
		},
		{"another seed used in cosmos tests #1", []byte{0}},
		{"another seed used in cosmos tests #2", []byte("mySecret")},
		{"another seed used in cosmos tests #3", []byte("")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotPrivKey, err := GenPrivKeyFromSecret(SECP256K1, tt.secret)
			require.NotNil(t, gotPrivKey)
			require.NoError(t, err)
			// interpret as a big.Int and make sure it is a valid field element:
			fe := new(big.Int).SetBytes(gotPrivKey.Bytes())
			require.True(t, fe.Cmp(N) < 0)
			require.True(t, fe.Sign() > 0)
		})
	}
}
