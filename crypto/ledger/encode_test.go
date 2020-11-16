package ledger_test

import (
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	hd "github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/ledger"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type byter interface {
	Bytes() []byte
}

func checkProtoJSON(t *testing.T, src proto.Message, dst proto.Message) {
	cdc := simapp.MakeTestEncodingConfig().Marshaler

	// Marshal to JSON bytes.
	js, err := cdc.MarshalJSON(src)
	require.Nil(t, err, "%+v", err)
	require.Contains(t, string(js), `"A0/vnNfExjWI07A/61KBudIyy6NNbz1xruWSEf+/4f6H"`) // The pubkey bytes as base64.
	// Unmarshal.
	err = cdc.UnmarshalJSON(js, dst)
	require.Nil(t, err, "%+v", err)
}

func TestEncodings(t *testing.T) {
	// Check PrivKey.
	path := hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv1, err := ledger.NewPrivKeySecp256k1Unsafe(path)
	require.NoError(t, err)
	var priv2 ledger.PrivKey
	checkProtoJSON(t, priv1, &priv2)
	require.EqualValues(t, priv1, &priv2)

	// Check PubKey.
	pub1 := priv1.PubKey()
	var pub2 secp256k1.PubKey
	checkProtoJSON(t, pub1, &pub2)
	require.EqualValues(t, pub1, &pub2)
}
