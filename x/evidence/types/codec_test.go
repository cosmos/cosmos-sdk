package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

func TestCodec(t *testing.T) {
	app := simapp.Setup(false)
	appCodec := codecstd.NewAppCodec(app.Codec())
	pk := ed25519.GenPrivKey()

	var e exported.Evidence = &types.Equivocation{
		Height:           10,
		Time:             time.Now().UTC(),
		Power:            100000,
		ConsensusAddress: pk.PubKey().Address().Bytes(),
	}
	bz, err := appCodec.MarshalEvidence(e)
	require.NoError(t, err)

	other, err := appCodec.UnmarshalEvidence(bz)
	require.NoError(t, err)
	require.Equal(t, e, other)
}
