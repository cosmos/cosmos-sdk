package transient_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	pruningTypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/store/transient"
	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"
)

var k, v = []byte("hello"), []byte("world")

func TestTransientStore(t *testing.T) {
	tstore := transient.NewStore()

	require.Nil(t, tstore.Get(k))

	tstore.Set(k, v)

	require.Equal(t, v, tstore.Get(k))

	tstore.Commit()

	require.Nil(t, tstore.Get(k))

	// no-op
	tstore.SetPruning(pruningTypes.NewPruningOptions(pruningTypes.PruningUndefined))

	emptyCommitID := tstore.LastCommitID()
	require.Equal(t, emptyCommitID.Version, int64(0))
	require.True(t, bytes.Equal(emptyCommitID.Hash, nil))
	require.Equal(t, types.StoreTypeTransient, tstore.GetStoreType())
}
