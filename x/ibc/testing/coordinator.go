package testing

import (
	"fmt"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

var (
	ChainIDPrefix   = "testchain"
	globalStartTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	timeIncrement   = time.Second * 5
)

// Coordinator is a testing struct which contains N TestChain's. It handles keeping all chains
// in sync with regards to time.
type Coordinator struct {
	Chains map[string]*TestChain
}

// NewCoordinator initializes Coordinator with N TestChain's
func NewCoordinator(t *testing.T, n int) *Coordinator {
	chains := make(map[string]*TestChain)

	for i := 0; i < n; i++ {
		chainID := ChainIDPrefix + string(i)
		chains[chainID] = NewTestChain(t, chainID)
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
func (coord *Coordinator) CommitBlock(chains ...string) {
	for _, chainID := range chains {
		chain := coord.Chains[chainID]
		chain.App.Commit()
		chain.NextBlock()
	}
	coord.IncrementTime()
}

// CommitNBlocks commits n blocks to state and updates the block height by 1 for each commit.
func (coord *Coordinator) CommitNBlocks(chainID string, n uint64) {
	chain := coord.Chains[chainID]

	for i := uint64(0); i < n; i++ {
		chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})
		chain.App.Commit()
		chain.NextBlock()
		coord.IncrementTime()
	}
}

// CreateClient creates a counterparty client on the source chain and returns the clientID.
func (coord *Coordinator) CreateClient(
	sourceID, counterpartyID string,
	clientType clientexported.ClientType,
) (clientID string, err error) {
	coord.CommitBlock(sourceID, counterpartyID)

	source := coord.Chains[sourceID]
	counterparty := coord.Chains[counterpartyID]

	clientID = source.NewClientID(counterparty.ChainID)

	switch clientType {
	case clientexported.Tendermint:
		err = source.CreateTMClient(counterparty, clientID)

	default:
		err = fmt.Errorf("client type %s is not supported", clientType)
	}

	if err != nil {
		return "", err
	}

	coord.IncrementTime()

	return clientID, nil
}

// UpdateClient updates a counterparty client on the source chain.
func (coord *Coordinator) UpdateClient(
	sourceID, counterpartyID,
	clientID string,
	clientType clientexported.ClientType,
) (err error) {
	coord.CommitBlock(sourceID, counterpartyID)

	source := coord.Chains[sourceID]
	counterparty := coord.Chains[counterpartyID]

	switch clientType {
	case clientexported.Tendermint:
		err = source.UpdateTMClient(counterparty, clientID)

	default:
		err = fmt.Errorf("client type %s is not supported", clientType)
	}

	if err != nil {
		return err
	}

	coord.IncrementTime()

	return nil
}

// CreateConnection constructs and executes connection handshake messages in order to create
// OPEN channels on source and counterparty chains. The connection information of the source
// and counterparty's are returned within a TestConnection struct. If there is a fault in
// the connection handshake then an error is returned.
//
// NOTE: The counterparty testing connection will be created even if it is not created in the
// application state.
func (coord *Coordinator) CreateConnection(
	sourceID, counterpartyID,
	clientID, counterpartyClientID string,
	state connectiontypes.State,
) (TestConnection, TestConnection, error) {
	source := coord.Chains[sourceID]
	counterparty := coord.Chains[counterpartyID]

	sourceConnection := source.NewTestConnection(clientID, counterpartyClientID)
	counterpartyConnection := counterparty.NewTestConnection(counterpartyClientID, clientID)

	if err := coord.CreateConnectionInit(source, counterparty, sourceConnection, counterpartyConnection); err != nil {
		return sourceConnection, counterpartyConnection, err
	}

	if err := coord.CreateConnectionOpenTry(counterparty, source, counterpartyConnection, sourceConnection); err != nil {
		return sourceConnection, counterpartyConnection, err
	}

	if err := coord.CreateConnectionOpenAck(source, counterparty, sourceConnection, counterpartyConnection); err != nil {
		return sourceConnection, counterpartyConnection, err
	}

	if err := coord.CreateConnectionOpenConfirm(counterparty, source, counterpartyConnection, sourceConnection); err != nil {
		return sourceConnection, counterpartyConnection, err
	}

	return sourceConnection, counterpartyConnection, nil
}

// CreateConenctionInit initializes a connection on the source chain with the state INIT
func (coord *Coordinator) CreateConnectionInit(
	source, counterparty *TestChain,
	sourceConnection, counterpartyConnection TestConnection,
) error {
	// initialize connection on source
	if err := source.ConnectionOpenInit(counterparty, sourceConnection, counterpartyConnection); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty.ChainID, source.ChainID,
		counterpartyConnection.ClientID, clientexported.Tendermint,
	); err != nil {
		return err
	}

	return nil
}

// CreateConenctionOpenTry initializes a connection on the source chain with the state TRYOPEN.
func (coord *Coordinator) CreateConnectionOpenTry(
	source, counterparty *TestChain,
	sourceConnection, counterpartyConnection TestConnection,
) error {
	// initialize TRYOPEN connection on source
	if err := source.ConnectionOpenTry(counterparty, sourceConnection, counterpartyConnection); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty.ChainID, source.ChainID,
		counterpartyConnection.ClientID, clientexported.Tendermint,
	); err != nil {
		return err
	}

	return nil
}

// CreateConnectionOpenAck initializes a connection on the source chain with the state OPEN
// using the OpenAck handshake call.
func (coord *Coordinator) CreateConnectionOpenAck(
	source, counterparty *TestChain,
	sourceConnection, counterpartyConnection TestConnection,
) error {
	// set OPEN connection on source using OpenAck
	if err := source.ConnectionOpenAck(counterparty, sourceConnection, counterpartyConnection); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty.ChainID, source.ChainID,
		counterpartyConnection.ClientID, clientexported.Tendermint,
	); err != nil {
		return err
	}

	return nil
}

// CreateConnectionOpenConfirm initializes a connection on the source chain with the state OPEN
// using the OpenConfirm handshake call.
func (coord *Coordinator) CreateConnectionOpenConfirm(
	source, counterparty *TestChain,
	sourceConnection, counterpartyConnection TestConnection,
) error {
	if err := counterparty.ConnectionOpenConfirm(counterparty, sourceConnection, counterpartyConnection); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		source.ChainID, counterparty.ChainID,
		sourceConnection.ClientID, clientexported.Tendermint,
	); err != nil {
		return err
	}

	return nil
}

// CreateChannel constructs and executes channel handshake messages in order to create
// channels on source and counterparty chains with the passed in Channel State. The portID and
// channelID of source and counterparty are returned.
//
// NOTE: The counterparty testing channel will be created even if it is not created in the
// application state.
func (coord *Coordinator) CreateChannel(
	sourceID, counterpartyID string,
	connection, counterpartyConnection TestConnection,
	order channeltypes.Order,
	state channeltypes.State,
) (sourceChannel, counterpartyChannel TestChannel) {
	source := coord.Chains[sourceID]
	counterparty := coord.Chains[counterpartyID]

	if state == channeltypes.UNINITIALIZED {
		return
	}

	sourceChannel = source.NewTestChannel()
	counterpartyChannel = counterparty.NewTestChannel()

	// Initialize channel on source
	source.ChannelOpenInit(sourceChannel, counterpartyChannel, order, connection.ID)
	coord.IncrementTime()

	// update counterparty client on source
	coord.UpdateClient(sourceID, counterpartyID, connection.ID, clientexported.Tendermint)

	if state == channeltypes.INIT {
		return
	}

	// Initialize channel on counterparty
	counterparty.ChannelOpenTry(counterpartyChannel, sourceChannel, order, counterpartyConnection.ID)
	coord.IncrementTime()

	// update source client on counterparty
	coord.UpdateClient(counterpartyID, sourceID, counterpartyConnection.ID, clientexported.Tendermint)

	if state == channeltypes.TRYOPEN {
		return
	}

	// Open both channel ends
	source.ChannelOpenAck(sourceChannel, counterpartyChannel, connection.ID)
	coord.IncrementTime()

	// update counterparty client on source
	coord.UpdateClient(sourceID, counterpartyID, connection.ID, clientexported.Tendermint)

	counterparty.ChannelOpenConfirm(counterpartyChannel, sourceChannel, counterpartyConnection.ID)
	coord.IncrementTime()

	// update source client on counterparty
	coord.UpdateClient(counterpartyID, sourceID, counterpartyConnection.ID, clientexported.Tendermint)

	return sourceChannel, counterpartyChannel
}
