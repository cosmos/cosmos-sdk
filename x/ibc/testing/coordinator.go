package testing

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// TODO: change funcs to take in test chain pointer instead of id string

var (
	ChainIDPrefix   = "testchain"
	globalStartTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	timeIncrement   = time.Second * 5
)

// Coordinator is a testing struct which contains N TestChain's. It handles keeping all chains
// in sync with regards to time.
type Coordinator struct {
	t *testing.T

	Chains map[string]*TestChain
}

// NewCoordinator initializes Coordinator with N TestChain's
func NewCoordinator(t *testing.T, n int) *Coordinator {
	chains := make(map[string]*TestChain)

	for i := 0; i < n; i++ {
		chainID := GetChainID(i)
		chains[chainID] = NewTestChain(t, chainID)
	}
	return &Coordinator{
		t:      t,
		Chains: chains,
	}
}

// Setup constructs a TM client, connection, and channel on both chains provided. It fails if
// any error occurs. The clientID's and TestConnections are returned for both chains.
func (coord *Coordinator) Setup(
	source, counterparty *TestChain,
) (string, string, *TestConnection, *TestConnection) {
	sourceClient, counterpartyClient, sourceConn, counterpartyConn := coord.SetupClientConnections(source, counterparty, clientexported.Tendermint)

	// channels can be referenced through the returned connections
	coord.CreateChannel(source, counterparty, sourceConn, counterpartyConn, channeltypes.UNORDERED)

	return sourceClient, counterpartyClient, sourceConn, counterpartyConn
}

// SetupClients is a helper function to create clients  on both the source and counterparty
// chain. It assumes the caller does not anticipate any errors.
func (coord *Coordinator) SetupClients(
	source, counterparty *TestChain,
	clientType clientexported.ClientType,
) (string, string) {

	sourceClient, err := coord.CreateClient(source, counterparty, clientType)
	require.NoError(coord.t, err)

	counterpartyClient, err := coord.CreateClient(counterparty, source, clientType)
	require.NoError(coord.t, err)

	return sourceClient, counterpartyClient
}

// SetupClientConnections is a helper function to create clients and the appropriate
// connections on both the source and counterparty chain. It assumes the caller does not
// anticipate any errors.
func (coord *Coordinator) SetupClientConnections(
	source, counterparty *TestChain,
	clientType clientexported.ClientType,
) (string, string, *TestConnection, *TestConnection) {

	sourceClient, counterpartyClient := coord.SetupClients(source, counterparty, clientType)

	sourceConnection, counterpartyConnection := coord.CreateConnection(source, counterparty, sourceClient, counterpartyClient)

	return sourceClient, counterpartyClient, sourceConnection, counterpartyConnection
}

// CreateClient creates a counterparty client on the source chain and returns the clientID.
func (coord *Coordinator) CreateClient(
	source, counterparty *TestChain,
	clientType clientexported.ClientType,
) (clientID string, err error) {
	coord.CommitBlock(source, counterparty)

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
	source, counterparty *TestChain,
	clientID string,
	clientType clientexported.ClientType,
) (err error) {
	coord.CommitBlock(source, counterparty)

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
// and counterparty's are returned within a TestConnection struct. The function expects the
// connections to be successfully created and testing will fail otherwise.
//
// NOTE: The counterparty testing connection will be created even if it is not created in the
// application state.
func (coord *Coordinator) CreateConnection(
	source, counterparty *TestChain,
	clientID, counterpartyClientID string,
) (*TestConnection, *TestConnection) {

	sourceConnection, counterpartyConnection, err := coord.CreateConnectionInit(source, counterparty, clientID, counterpartyClientID)
	require.NoError(coord.t, err)

	err = coord.CreateConnectionOpenTry(counterparty, source, counterpartyConnection, sourceConnection)
	require.NoError(coord.t, err)

	err = coord.CreateConnectionOpenAck(source, counterparty, sourceConnection, counterpartyConnection)
	require.NoError(coord.t, err)

	err = coord.CreateConnectionOpenConfirm(counterparty, source, counterpartyConnection, sourceConnection)
	require.NoError(coord.t, err)

	return sourceConnection, counterpartyConnection
}

// CreateChannel constructs and executes channel handshake messages in order to create
// channels on source and counterparty chains with the passed in Channel State. The portID
// and channelID of source and counterparty are returned. The function expects the creation
// of both channels to succeed and testing fails otherwise.
//
// NOTE: The counterparty testing channel will be created even if it is not created in the
// application state.
func (coord *Coordinator) CreateChannel(
	source, counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
	order channeltypes.Order,
) (TestChannel, TestChannel) {

	sourceChannel, counterpartyChannel, err := coord.CreateChannelOpenInit(source, counterparty, connection, counterpartyConnection, order)
	require.NoError(coord.t, err)

	err = coord.CreateChannelOpenTry(counterparty, source, counterpartyChannel, sourceChannel, counterpartyConnection, order)
	require.NoError(coord.t, err)

	err = coord.CreateChannelOpenAck(source, counterparty, sourceChannel, counterpartyChannel, connection)
	require.NoError(coord.t, err)

	err = coord.CreateChannelOpenConfirm(counterparty, source, counterpartyChannel, sourceChannel, counterpartyConnection)
	require.NoError(coord.t, err)

	return sourceChannel, counterpartyChannel
}

// IncrementTime iterates through all the TestChain's and increments their current header time
// by 5 seconds.
//
// CONTRACT: this function must be called after every commit on any TestChain.
func (coord *Coordinator) IncrementTime() {
	for _, chain := range coord.Chains {
		chain.CurrentHeader.Time = chain.CurrentHeader.Time.Add(timeIncrement)
		chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})
	}
}

// GetChain returns the TestChain using the given chainID and returns an error if it does
// not exist.
func (coord *Coordinator) GetChain(chainID string) *TestChain {
	chain, found := coord.Chains[chainID]
	require.True(coord.t, found, fmt.Sprintf("%s chain does not exist", chainID))
	return chain
}

// GetChainID returns the chainID used for the provided index.
func GetChainID(index int) string {
	return ChainIDPrefix + strconv.Itoa(index)
}

// CommitBlock commits a block on the provided indexes and then increments the global time.
//
// CONTRACT: the passed in list of indexes must not contain duplicates
func (coord *Coordinator) CommitBlock(chains ...*TestChain) {
	for _, chain := range chains {
		chain.App.Commit()
		chain.NextBlock()
	}
	coord.IncrementTime()
}

// CommitNBlocks commits n blocks to state and updates the block height by 1 for each commit.
func (coord *Coordinator) CommitNBlocks(chain *TestChain, n uint64) {
	for i := uint64(0); i < n; i++ {
		chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})
		chain.App.Commit()
		chain.NextBlock()
		coord.IncrementTime()
	}
}

// CreateConenctionInit initializes a connection on the source chain with the state INIT
// using the OpenInit handshake call.
//
// NOTE: The counterparty testing connection will be created even if it is not created in the
// application state.
func (coord *Coordinator) CreateConnectionInit(
	source, counterparty *TestChain,
	clientID, counterpartyClientID string,
) (*TestConnection, *TestConnection, error) {
	sourceConnection := source.NewTestConnection(clientID, counterpartyClientID)
	counterpartyConnection := counterparty.NewTestConnection(counterpartyClientID, clientID)

	// initialize connection on source
	if err := source.ConnectionOpenInit(counterparty, sourceConnection, counterpartyConnection); err != nil {
		return sourceConnection, counterpartyConnection, err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty, source,
		counterpartyClientID, clientexported.Tendermint,
	); err != nil {
		return sourceConnection, counterpartyConnection, err
	}

	return sourceConnection, counterpartyConnection, nil
}

// CreateConenctionOpenTry initializes a connection on the source chain with the state TRYOPEN
// using the OpenTry handshake call.
func (coord *Coordinator) CreateConnectionOpenTry(
	source, counterparty *TestChain,
	sourceConnection, counterpartyConnection *TestConnection,
) error {
	// initialize TRYOPEN connection on source
	if err := source.ConnectionOpenTry(counterparty, sourceConnection, counterpartyConnection); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty, source,
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
	sourceConnection, counterpartyConnection *TestConnection,
) error {
	// set OPEN connection on source using OpenAck
	if err := source.ConnectionOpenAck(counterparty, sourceConnection, counterpartyConnection); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty, source,
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
	sourceConnection, counterpartyConnection *TestConnection,
) error {
	if err := source.ConnectionOpenConfirm(counterparty, sourceConnection, counterpartyConnection); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty, source,
		counterpartyConnection.ClientID, clientexported.Tendermint,
	); err != nil {
		return err
	}

	return nil
}

// CreateChannelOpenInit initializes a channel on the source chain with the state INIT
// using the OpenInit handshake call.
//
// NOTE: The counterparty testing channel will be created even if it is not created in the
// application state.
func (coord *Coordinator) CreateChannelOpenInit(
	source, counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
	order channeltypes.Order,
) (TestChannel, TestChannel, error) {
	sourceChannel := connection.AddTestChannel()
	counterpartyChannel := counterpartyConnection.AddTestChannel()

	// create port capability
	source.CreatePortCapability(sourceChannel.PortID)
	coord.IncrementTime()

	// initialize channel on source
	if err := source.ChannelOpenInit(sourceChannel, counterpartyChannel, order, connection.ID); err != nil {
		return sourceChannel, counterpartyChannel, err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty, source,
		counterpartyConnection.ClientID, clientexported.Tendermint,
	); err != nil {
		return sourceChannel, counterpartyChannel, err
	}

	return sourceChannel, counterpartyChannel, nil
}

// CreateChannelOpenTry initializes a channel on the source chain with the state TRYOPEN
// using the OpenTry handshake call.
func (coord *Coordinator) CreateChannelOpenTry(
	source, counterparty *TestChain,
	sourceChannel, counterpartyChannel TestChannel,
	connection *TestConnection,
	order channeltypes.Order,
) error {

	// initialize channel on source
	if err := source.ChannelOpenTry(counterparty, sourceChannel, counterpartyChannel, order, connection.ID); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty, source,
		connection.CounterpartyClientID, clientexported.Tendermint,
	); err != nil {
		return err
	}

	return nil
}

// CreateChannelOpenAck initializes a channel on the source chain with the state OPEN
// using the OpenAck handshake call.
func (coord *Coordinator) CreateChannelOpenAck(
	source, counterparty *TestChain,
	sourceChannel, counterpartyChannel TestChannel,
	connection *TestConnection,
) error {

	// initialize channel on source
	if err := source.ChannelOpenAck(counterparty, sourceChannel, counterpartyChannel); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty, source,
		connection.CounterpartyClientID, clientexported.Tendermint,
	); err != nil {
		return err
	}

	return nil
}

// CreateChannelOpenConfirm initializes a channel on the source chain with the state OPEN
// using the OpenConfirm handshake call.
func (coord *Coordinator) CreateChannelOpenConfirm(
	source, counterparty *TestChain,
	sourceChannel, counterpartyChannel TestChannel,
	connection *TestConnection,
) error {

	// initialize channel on source
	if err := source.ChannelOpenConfirm(counterparty, sourceChannel, counterpartyChannel); err != nil {
		return err
	}
	coord.IncrementTime()

	// update source client on counterparty connection
	if err := coord.UpdateClient(
		counterparty, source,
		connection.CounterpartyClientID, clientexported.Tendermint,
	); err != nil {
		return err
	}

	return nil
}
