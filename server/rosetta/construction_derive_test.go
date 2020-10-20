package rosetta

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/require"

	cosmos "github.com/cosmos/cosmos-sdk/server/rosetta/client/sdk"
	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"
)

func TestLaunchpad_ConstructionDerive(t *testing.T) {
	data, err := hex.DecodeString("037521798512c0ebde2b79f5c72121fdd652ee9482e6973d507473b3ff720a3bdf")
	require.NoError(t, err)

	properties := properties{
		Blockchain: "TheBlockchain",
		Network:    "TheNetwork",
		AddrPrefix: "test",
	}

	adapter := newAdapter(cosmos.NewClient(""), tendermint.NewClient(""), properties)
	deriveResp, deriveErr := adapter.ConstructionDerive(context.Background(), &types.ConstructionDeriveRequest{
		PublicKey: &types.PublicKey{
			Bytes:     data,
			CurveType: "secp256k1",
		},
	})
	require.Nil(t, deriveErr)
	require.NotNil(t, deriveResp)
	require.Equal(t, "test1f0ykmmcx9cwjlk532w4lvkac875wjrj9w7538z", deriveResp.Address)

	// TODO: Use table driven tests
	// check unsupported curve returns error
	_, deriveErr = adapter.ConstructionDerive(context.Background(), &types.ConstructionDeriveRequest{
		PublicKey: &types.PublicKey{
			Bytes:     data,
			CurveType: "edwards25519",
		},
	})
	require.Equal(t, ErrUnsupportedCurve, deriveErr)
}
