package testing

import (
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

var (
	ClientIDPrefix  = "clientIDForChain"
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

	for i, chain := range suite.Chains {
		chain = NewTestChain(suite.T(), ClientIDPrefix+string(i))
	}
}

// IncrementTime iterates through all the TestChain's and increments their current header time
// by 5 seconds.
//
// CONTRACT: this function must be called after every commit on any TestChain.
func (suite *IBCTestSuite) IncrementTime() {
	for _, chain := range suite.Chains {
		chain.CurrentHeader = abci.Header{
			Height: chain.CurrentHeader.Height,
			Time:   chain.CurrentHeader.Time.Add((timeIncrement)),
		}
	}
}

// CommitBlock commits a block on the provided indexes and then increments the global time.
//
// CONTRACT: the passed in list of indexes must not contain duplicates
func (suite *IBCTestSuite) CommitBlock(chains ...uint64) {
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

	for i := uint64(0); i < n; i++ {
		chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})
		chain.App.Commit()
		nextBlock(chain)
		suite.IncrementTime()
	}
}

// CreateClient creates a counterparty client on the source chain.
func (suite *IBCTestSuite) CreateClient(source, counterparty uint64) {
	suite.CommitBlock(source, counterparty)
	suite.Chains[source].CreateClient(suite.Chains[counterparty])
	suite.IncrementTime()
}

// UpdateClient updates a counterparty client on the source chain.
func (suite *IBCTestSuite) UpdateClient(source, counterparty uint64) {
	suite.CommitBlock(source, counterparty)
	suite.Chains[source].UpdateClient(suite.Chains[counterparty])
	suite.IncrementTime()
}

// CreateConnection constructs and executes connection handshake messages in order to create
// channels on source and counterparty chains with the passed in Connection State. The
// connectionID's of the source and counterparty are returned.
//
// NOTE: The counterparty testing connection will be created even if it is not created in the
// application state.
func (suite *IBCTestSuite) CreateConnection(
	sourceIndex, counterpartyIndex uint64,
	state connectiontypes.State,
) (sourceConnection, counterpartyConnection string) {
	source := suite.Chains[sourceIndex]
	counterparty := suite.Chains[counterpartyIndex]

	if state == connectiontypes.UNINITIALIZED {
		return
	}

	sourceConnection = source.NewConnection()
	counterpartyConnection = counterparty.NewConnection()

	// Initialize connection on source
	source.ConnectionOpenInit(counterparty, sourceConnection, counterpartyConnection)
	suite.IncrementTime()

	if state == connectiontypes.INIT {
		return
	}

	// Initialize connection on counterparty
	counterparty.ConnectionOpenTry(source, counterpartyConnection, sourceConnection)
	suite.IncrementTime()

	if state == connectiontypes.TRYOPEN {
		return
	}

	// Open connection on both chains
	source.ConnectionOpenAck(counterparty, sourceConnection, counterpartyConnection)
	suite.IncrementTime()

	counterparty.ConnectionOpenConfirm(source, counterpartyConnection, sourceConnection)
	suite.IncrementTime()
	return sourceConnection, counterpartyConnection
}

// CreateChannel constructs and executes channel handshake messages in order to create
// channels on source and counterparty chains with the passed in Channel State. The portID and
// channelID of source and counterparty are returned.
//
// NOTE: The counterparty testing channel will be created even if it is not created in the
// application state.
func (suite *IBCTestSuite) CreateChannel(
	sourceIndex, counterpartyIndex,
	connection uint64,
	order channeltypes.Order,
	state channeltypes.State,
) (sourceChannel, counterpartyChannel Channel) {
	source := suite.Chains[sourceIndex]
	counterparty := suite.Chains[counterpartyIndex]

	if state == channeltypes.UNINITIALIZED {
		return
	}

	sourceChannel = source.NewChannel()
	counterpartyChannel = counterparty.NewChannel()

	connectionID := source.Connections[connection]

	// Initialize channel on source
	source.ChannelOpenInit(sourceChannel, counterpartyChannel, order, connectionID)
	suite.IncrementTime()

	if state == channeltypes.INIT {
		return
	}

	// Initialize channel on counterparty
	counterparty.ChannelOpenTry(counterpartyChannel, sourceChannel, order, connectionID)
	suite.IncrementTime()

	if state == channeltypes.TRYOPEN {
		return
	}

	// Open both channel ends
	source.ChannelOpenAck(sourceChannel, counterpartyChannel, connectionID)
	suite.IncrementTime()

	counterparty.ChannelOpenConfirm(counterpartyChannel, sourceChannel, connectionID)
	suite.IncrementTime()
	return sourceChannel, counterpartyChannel
}
