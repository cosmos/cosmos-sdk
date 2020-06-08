package testing

import (
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	ClientID        = "clientIDForChain"
	globalStartTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	timeIncrement   = time.Second * 5
)

// IBCTestSuite is a testing struct which wraps the testing suite with N TestChain's. It handles
// keeping all chains in time-sync.
type IBCTestSuite struct {
	suite.Suite

	Chains []*TestChain
}

// SetupTest initializes IBCTestSuite with N TestChain's
func (suite *IBCTestSuite) SetupTest(n uint64) {
	suite.Chains = make([]*TestChain, n)

	for i, _ := range suite.Chains {
		suite.Chains[i] = NewTestChain(suite.T(), clientID+string(i))
	}
}

// IncrementTime iterates through all the TestChain's and increments their current header time
// by 5 seconds.
//
// CONTRACT: this function must be called after every commit on any TestChain.
func (suite *IBCTestSuite) IncrementTime() {
	for _, chain := range suite.Chains {
		chain.UpdateBlockTime()
	}
}

// CommitBlock commits a block on the provided indexes and then increments the global time.
//
// CONTRACT: the passed in list of indexes must not contain duplicates
func (suite *IBCKeeperSuite) CommitBlock(chains []uint64) {
	for _, index := range chains {
		chain := suite.Chains[index]
		chain.App.Commit()
		nextBlock(chain)
	}
	suite.IncrementTime()
}

// CommitNBlocks commits n blocks to state and updates the block height by 1 for each commit.
func (suite *IBCTestSuite) CommitNBlocks(index, n uint64) {
	chain := suite.Chains[index]

	for i := 0; i < n; i++ {
		chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})
		chain.App.Commit()
		nextBlock(chain)
		suite.IncrementTime()
	}
}

// CreateClient creates a counterparty client on the source chain.
func (suite *IBCTestSuite) CreateClient(source, counterparty uint64) {
	suite.CommitBlock(source, counterparty...)
	suite.Chains[source].CreateClient(suite.Chains[counterparty])
	suite.IncrementTime()
}

// UpdateClient updates a counterparty client on the source chain.
func (suite *IBCTestSuite) UpdateClient(source, counterparty uint64) {
	commitBlock(source, counterparty...)
	suite.Chains[source].UpdateClient(suite.Chains[counterparty])
	suite.IncrementTime()
}
