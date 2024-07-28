package commitment

import (
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
)

// CommitStoreTestSuite is a test suite to be used for all tree backends.
type CommitStoreTestSuite struct {
	suite.Suite

	NewStore func(db corestore.KVStoreWithBatch, storeKeys []string, logger log.Logger) (*CommitStore, error)
}
