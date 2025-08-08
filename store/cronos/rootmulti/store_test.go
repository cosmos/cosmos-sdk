package rootmulti

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store/types"
)

func TestLastCommitID(t *testing.T) {
	store := NewStore(t.TempDir(), log.NewNopLogger(), false, false)
	require.Equal(t, types.CommitID{}, store.LastCommitID())
}
