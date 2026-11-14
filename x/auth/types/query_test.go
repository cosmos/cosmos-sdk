package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestQueryPubKeyHistoryResponse_UnpackInterfaces(t *testing.T) {
	pk := secp256k1.GenPrivKey().PubKey()
	anyPk, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)

	bz, err := (&types.QueryPubKeyHistoryResponse{
		Entries: []types.PubKeyHistoryEntry{{PubKey: anyPk, ReplacedAtHeight: 5}},
	}).Marshal()
	require.NoError(t, err)

	// Decoding from bytes leaves the Any cached values empty until unpacked.
	var resp types.QueryPubKeyHistoryResponse
	require.NoError(t, resp.Unmarshal(bz))
	require.Nil(t, resp.Entries[0].PubKey.GetCachedValue())

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	require.NoError(t, codectypes.UnpackInterfaces(&resp, registry))

	cached, ok := resp.Entries[0].PubKey.GetCachedValue().(cryptotypes.PubKey)
	require.True(t, ok, "entry pubkey was not unpacked")
	require.True(t, pk.Equals(cached))
}
