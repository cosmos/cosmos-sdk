package legacybech32

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pointnetwork/cosmos-point-sdk/crypto/hd"
	"github.com/pointnetwork/cosmos-point-sdk/crypto/ledger"
	"github.com/pointnetwork/cosmos-point-sdk/testutil/testdata"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
)

func TestBeach32ifPbKey(t *testing.T) {
	require := require.New(t)
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, err := ledger.NewPrivKeySecp256k1Unsafe(path)
	require.Nil(err, "%s", err)
	require.NotNil(priv)

	pubKeyAddr, err := MarshalPubKey(AccPK, priv.PubKey())
	require.NoError(err)
	require.Equal("cosmospub1addwnpepqd87l8xhcnrrtzxnkql7k55ph8fr9jarf4hn6udwukfprlalu8lgw0urza0",
		pubKeyAddr, "Is your device using test mnemonic: %s ?", testdata.TestMnemonic)
}
