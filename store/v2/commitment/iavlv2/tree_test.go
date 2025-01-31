package iavlv2

import (
	"fmt"
	"testing"

	"github.com/cosmos/iavl/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corelog "cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2/commitment"
)

func TestCommitterSuite(t *testing.T) {
	s := &commitment.CommitStoreTestSuite{
		TreeType: "iavlv2",
		NewStore: func(
			db corestore.KVStoreWithBatch,
			dbDir string,
			storeKeys, oldStoreKeys []string,
			logger corelog.Logger,
		) (*commitment.CommitStore, error) {
			multiTrees := make(map[string]commitment.Tree)
			mountTreeFn := func(storeKey string) (commitment.Tree, error) {
				path := fmt.Sprintf("%s/%s", dbDir, storeKey)
				tree, err := NewTree(DefaultConfig(), iavl.SqliteDbOptions{Path: path}, logger)
				require.NoError(t, err)
				return tree, nil
			}
			for _, storeKey := range storeKeys {
				multiTrees[storeKey], _ = mountTreeFn(storeKey)
			}
			oldTrees := make(map[string]commitment.Tree)
			for _, storeKey := range oldStoreKeys {
				oldTrees[storeKey], _ = mountTreeFn(storeKey)
			}

			return commitment.NewCommitStore(multiTrees, oldTrees, db, logger)
		},
	}

	suite.Run(t, s)
}
