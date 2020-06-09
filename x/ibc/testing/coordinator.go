package testing

import (
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

var (
	ClientIDPrefix  = "clientIDForChain"
	globalStartTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	timeIncrement   = time.Second * 5
)

// Coordinator is a testing struct which contains N TestChain's. It handles keeping all chains
// in sync with regards to time.
type Coordinator struct {
	Chains []*TestChain
}

// NewCoordinator initializes Coordinator with N TestChain's
func NewCoordinator(t *testing.T, n uint64) *Coordinator {
	chains := make([]*TestChain, n)

	for i := range chains {
		chains[i] = NewTestChain(t, ClientIDPrefix+string(i))
	}
	return &Coordinator{
		Chains: chains,
	}
}

// IncrementTime iterates through all the TestChain's and increments their current header time
// by 5 seconds.
//
// CONTRACT: this function must be called after every commit on any TestChain.
func (coord *Coordinator) IncrementTime() {
	for _, chain := range coord.Chains {
		chain.CurrentHeader = abci.Header{
			Height: chain.CurrentHeader.Height,
			Time:   chain.CurrentHeader.Time.Add((timeIncrement)),
		}
	}
}

// CommitBlock commits a block on the provided indexes and then increments the global time.
//
// CONTRACT: the passed in list of indexes must not contain duplicates
func (coord *Coordinator) CommitBlock(chains ...uint64) {
	for _, index := range chains {
		chain := coord.Chains[index]
		chain.App.Commit()
		nextBlock(chain)
	}
	coord.IncrementTime()
}

// CommitNBlocks commits n blocks to state and updates the block height by 1 for each commit.
func (coord *Coordinator) CommitNBlocks(index, n uint64) {
	chain := coord.Chains[index]

	for i := uint64(0); i < n; i++ {
		chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})
		chain.App.Commit()
		nextBlock(chain)
		coord.IncrementTime()
	}
}

// CreateClient creates a counterparty client on the source chain.
func (coord *Coordinator) CreateClient(source, counterparty uint64) {
	coord.CommitBlock(source, counterparty)
	coord.Chains[source].CreateClient(coord.Chains[counterparty])
	coord.IncrementTime()
}

// UpdateClient updates a counterparty client on the source chain.
func (coord *Coordinator) UpdateClient(source, counterparty uint64) {
	coord.CommitBlock(source, counterparty)
	coord.Chains[source].UpdateClient(coord.Chains[counterparty])
	coord.IncrementTime()
}

// CreateConnection constructs and executes connection handshake messages in order to create
// channels on source and counterparty chains with the passed in Connection State. The
// connectionID's of the source and counterparty are returned.
//
// NOTE: The counterparty testing connection will be created even if it is not created in the
// application state.
func (coord *Coordinator) CreateConnection(
	sourceIndex, counterpartyIndex uint64,
	state connectiontypes.State,
) (sourceConnection, counterpartyConnection string) {
	source := coord.Chains[sourceIndex]
	counterparty := coord.Chains[counterpartyIndex]

	if state == connectiontypes.UNINITIALIZED {
		return
	}

	sourceConnection = source.NewConnection()
	counterpartyConnection = counterparty.NewConnection()

	// Initialize connection on source
	source.ConnectionOpenInit(counterparty, sourceConnection, counterpartyConnection)
	coord.IncrementTime()

	if state == connectiontypes.INIT {
		return
	}

	// Initialize connection on counterparty
	counterparty.ConnectionOpenTry(source, counterpartyConnection, sourceConnection)
	coord.IncrementTime()

	if state == connectiontypes.TRYOPEN {
		return
	}

	// Open connection on both chains
	source.ConnectionOpenAck(counterparty, sourceConnection, counterpartyConnection)
	coord.IncrementTime()

	counterparty.ConnectionOpenConfirm(source, counterpartyConnection, sourceConnection)
	coord.IncrementTime()
	return sourceConnection, counterpartyConnection
}

// CreateChannel constructs and executes channel handshake messages in order to create
// channels on source and counterparty chains with the passed in Channel State. The portID and
// channelID of source and counterparty are returned.
//
// NOTE: The counterparty testing channel will be created even if it is not created in the
// application state.
func (coord *Coordinator) CreateChannel(
	sourceIndex, counterpartyIndex uint64,
	connectionID, counterpartyConnectionID string,
	order channeltypes.Order,
	state channeltypes.State,
) (sourceChannel, counterpartyChannel Channel) {
	source := coord.Chains[sourceIndex]
	counterparty := coord.Chains[counterpartyIndex]

	if state == channeltypes.UNINITIALIZED {
		return
	}

	sourceChannel = source.NewChannel()
	counterpartyChannel = counterparty.NewChannel()

	// Initialize channel on source
	source.ChannelOpenInit(sourceChannel, counterpartyChannel, order, connectionID)
	coord.IncrementTime()

	if state == channeltypes.INIT {
		return
	}

	// Initialize channel on counterparty
	counterparty.ChannelOpenTry(counterpartyChannel, sourceChannel, order, counterpartyConnectionID)
	coord.IncrementTime()

	if state == channeltypes.TRYOPEN {
		return
	}

	// Open both channel ends
	source.ChannelOpenAck(sourceChannel, counterpartyChannel, connectionID)
	coord.IncrementTime()

	counterparty.ChannelOpenConfirm(counterpartyChannel, sourceChannel, counterpartyConnectionID)
	coord.IncrementTime()
	return sourceChannel, counterpartyChannel
}
