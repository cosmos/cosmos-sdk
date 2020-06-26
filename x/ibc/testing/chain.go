package testing

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmmath "github.com/tendermint/tendermint/libs/math"
	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/keeper"
)

const (
	// Default params used to create a TM client
	TrustingPeriod  time.Duration = time.Hour * 24 * 7 * 2
	UnbondingPeriod time.Duration = time.Hour * 24 * 7 * 3
	MaxClockDrift   time.Duration = time.Second * 10

	ConnectionVersion = "1.0.0"
	ChannelVersion    = "ics20-1"
	InvalidID         = "IDisInvalid"

	ConnectionIDPrefix = "connectionid"

	maxInt = int(^uint(0) >> 1)
)

var (
	DefaultTrustLevel tmmath.Fraction = lite.DefaultTrustLevel
	TestHash                          = []byte("TESTING HASH")
)

// TestChain is a testing struct that wraps a simapp with the last TM Header, the current ABCI
// header and the validators of the TestChain. It also contains a field called ChainID. This
// is the clientID that *other* chains use to refer to this TestChain. The SenderAccount
// is used for delivering transactions through the application state.
// NOTE: the actual application uses an empty chain-id for ease of testing.
type TestChain struct {
	t *testing.T

	App           *simapp.SimApp
	ChainID       string
	LastHeader    ibctmtypes.Header // header for last block height committed
	CurrentHeader abci.Header       // header for current block height
	Querier       sdk.Querier

	Vals    *tmtypes.ValidatorSet
	Signers []tmtypes.PrivValidator

	senderPrivKey crypto.PrivKey
	SenderAccount authtypes.AccountI

	// IBC specific helpers
	ClientIDs   []string          // ClientID's used on this chain
	Connections []*TestConnection // track connectionID's created for this chain
}

// NewTestChain initializes a new TestChain instance with a single validator set using a
// generated private key. It also creates a sender account to be used for delivering transactions.
//
// The first block height is committed to state in order to allow for client creations on
// counterparty chains. The TestChain will return with a block height starting at 2.
//
// Time management is handled by the Coordinator in order to ensure synchrony between chains.
// Each update of any chain increments the block header time for all chains by 5 seconds.
func NewTestChain(t *testing.T, chainID string) *TestChain {
	// generate validator private/public key
	privVal := tmtypes.NewMockPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := []tmtypes.PrivValidator{privVal}

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
	}

	app := simapp.SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, balance)

	// create current header and call begin block
	header := abci.Header{
		Height: 1,
		Time:   globalStartTime,
	}

	// create an account to send transactions from
	chain := &TestChain{
		t:             t,
		ChainID:       chainID,
		App:           app,
		CurrentHeader: header,
		Querier:       keeper.NewQuerier(*app.IBCKeeper),
		Vals:          valSet,
		Signers:       signers,
		senderPrivKey: senderPrivKey,
		SenderAccount: acc,
		ClientIDs:     make([]string, 0),
		Connections:   make([]*TestConnection, 0),
	}

	chain.NextBlock()

	return chain
}

// GetContext returns the current context for the application.
func (chain *TestChain) GetContext() sdk.Context {
	return chain.App.BaseApp.NewContext(false, chain.CurrentHeader)
}

// QueryProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryProof(key []byte) ([]byte, uint64) {
	res := chain.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
		Height: chain.App.LastBlockHeight() - 1,
		Data:   key,
		Prove:  true,
	})

	merkleProof := commitmenttypes.MerkleProof{
		Proof: res.Proof,
	}

	proof, err := chain.App.AppCodec().MarshalBinaryBare(&merkleProof)
	require.NoError(chain.t, err)

	// proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree. Thus we return proof height + 1
	return proof, uint64(res.Height) + 1
}

// NextBlock sets the last header to the current header and increments the current header to be
// at the next block height. It does not update the time as that is handled by the Coordinator.
//
// CONTRACT: this function must only be called after app.Commit() occurs
func (chain *TestChain) NextBlock() {
	// set the last header to the current header
	chain.LastHeader = chain.CreateTMClientHeader()

	// increment the current header
	chain.CurrentHeader = abci.Header{
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time: chain.CurrentHeader.Time,
	}

	chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})

}

// SendMsg delivers a transaction through the application. It updates the senders sequence
// number and updates the TestChain's headers.
func (chain *TestChain) SendMsg(msg sdk.Msg) error {
	_, _, err := simapp.SignCheckDeliver(
		chain.t,
		chain.App.Codec(),
		chain.App.BaseApp,
		chain.GetContext().BlockHeader(),
		[]sdk.Msg{msg},
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		true, true, chain.senderPrivKey,
	)
	if err != nil {
		return err
	}

	// SignCheckDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)

	return nil
}

// GetClientState retreives the client state for the provided clientID. The client is
// expected to exist otherwise testing will fail.
func (chain *TestChain) GetClientState(clientID string) clientexported.ClientState {
	clientState, found := chain.App.IBCKeeper.ClientKeeper.GetClientState(chain.GetContext(), clientID)
	require.True(chain.t, found)

	return clientState
}

// GetConnection retreives an IBC Connection for the provided TestConnection. The
// connection is expected to exist otherwise testing will fail.
func (chain *TestChain) GetConnection(testConnection *TestConnection) connectiontypes.ConnectionEnd {
	connection, found := chain.App.IBCKeeper.ConnectionKeeper.GetConnection(chain.GetContext(), testConnection.ID)
	require.True(chain.t, found)

	return connection
}

// GetChannel retreives an IBC Channel for the provided TestChannel. The channel
// is expected to exist otherwise testing will fail.
func (chain *TestChain) GetChannel(testChannel TestChannel) channeltypes.Channel {
	channel, found := chain.App.IBCKeeper.ChannelKeeper.GetChannel(chain.GetContext(), testChannel.PortID, testChannel.ID)
	require.True(chain.t, found)

	return channel
}

// GetAcknowledgement retreives an acknowledgement for the provided packet. If the
// acknowledgement does not exist then testing will fail.
func (chain *TestChain) GetAcknowledgement(packet channelexported.PacketI) []byte {
	ack, found := chain.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	require.True(chain.t, found)

	return ack
}

// GetPrefix returns the prefix for used by a chain in connection creation
func (chain *TestChain) GetPrefix() commitmenttypes.MerklePrefix {
	return commitmenttypes.NewMerklePrefix(chain.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes())
}

// NewClientID appends a new clientID string in the format:
// ClientFor<counterparty-chain-id><index>
func (chain *TestChain) NewClientID(counterpartyChainID string) string {
	clientID := "client" + strconv.Itoa(len(chain.ClientIDs)) + "For" + counterpartyChainID
	chain.ClientIDs = append(chain.ClientIDs, clientID)
	return clientID
}

// NewConnection appends a new TestConnection which contains references to the connection id,
// client id and counterparty client id. The connection id format:
// connectionid<index>
func (chain *TestChain) NewTestConnection(clientID, counterpartyClientID string) *TestConnection {
	connectionID := ConnectionIDPrefix + strconv.Itoa(len(chain.Connections))
	conn := &TestConnection{
		ID:                   connectionID,
		ClientID:             clientID,
		CounterpartyClientID: counterpartyClientID,
	}

	chain.Connections = append(chain.Connections, conn)
	return conn
}

// CreateTMClient will construct and execute a 07-tendermint MsgCreateClient. A counterparty
// client will be created on the (target) chain.
func (chain *TestChain) CreateTMClient(counterparty *TestChain, clientID string) error {
	// construct MsgCreateClient using counterparty
	msg := ibctmtypes.NewMsgCreateClient(
		clientID, counterparty.LastHeader,
		DefaultTrustLevel, TrustingPeriod, UnbondingPeriod, MaxClockDrift,
		commitmenttypes.GetSDKSpecs(), chain.SenderAccount.GetAddress(),
	)

	return chain.SendMsg(msg)
}

// UpdateTMClient will construct and execute a 07-tendermint MsgUpdateClient. The counterparty
// client will be updated on the (target) chain.
func (chain *TestChain) UpdateTMClient(counterparty *TestChain, clientID string) error {
	msg := ibctmtypes.NewMsgUpdateClient(
		clientID, counterparty.LastHeader,
		chain.SenderAccount.GetAddress(),
	)

	return chain.SendMsg(msg)
}

// CreateTMClientHeader creates a TM header to update the TM client.
func (chain *TestChain) CreateTMClientHeader() ibctmtypes.Header {
	vsetHash := chain.Vals.Hash()
	tmHeader := tmtypes.Header{
		Version:            version.Consensus{Block: 2, App: 2},
		ChainID:            chain.ChainID,
		Height:             chain.CurrentHeader.Height,
		Time:               chain.CurrentHeader.Time,
		LastBlockID:        MakeBlockID(make([]byte, tmhash.Size), maxInt, make([]byte, tmhash.Size)),
		LastCommitHash:     chain.App.LastCommitID().Hash,
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: vsetHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            chain.CurrentHeader.AppHash,
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    chain.Vals.Proposer.Address,
	}
	hhash := tmHeader.Hash()

	blockID := MakeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))

	voteSet := tmtypes.NewVoteSet(chain.ChainID, chain.CurrentHeader.Height, 1, tmtypes.PrecommitType, chain.Vals)

	commit, err := tmtypes.MakeCommit(blockID, chain.CurrentHeader.Height, 1, voteSet, chain.Signers, chain.CurrentHeader.Time)
	require.NoError(chain.t, err)

	signedHeader := tmtypes.SignedHeader{
		Header: &tmHeader,
		Commit: commit,
	}

	return ibctmtypes.Header{
		SignedHeader: signedHeader,
		ValidatorSet: chain.Vals,
	}
}

// Copied unimported test functions from tmtypes to use them here
func MakeBlockID(hash []byte, partSetSize int, partSetHash []byte) tmtypes.BlockID {
	return tmtypes.BlockID{
		Hash: hash,
		PartsHeader: tmtypes.PartSetHeader{
			Total: partSetSize,
			Hash:  partSetHash,
		},
	}
}

// ConnectionOpenInit will construct and execute a MsgConnectionOpenInit.
func (chain *TestChain) ConnectionOpenInit(
	counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
) error {
	msg := connectiontypes.NewMsgConnectionOpenInit(
		connection.ID, connection.ClientID,
		counterpartyConnection.ID, connection.CounterpartyClientID,
		counterparty.GetPrefix(),
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ConnectionOpenTry will construct and execute a MsgConnectionOpenTry.
func (chain *TestChain) ConnectionOpenTry(
	counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
) error {
	connectionKey := host.KeyConnection(counterpartyConnection.ID)
	proofInit, proofHeight := counterparty.QueryProof(connectionKey)

	// retrieve consensus state to provide proof for
	consState, found := counterparty.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(counterparty.GetContext(), counterpartyConnection.ClientID)
	require.True(chain.t, found)

	consensusHeight := consState.GetHeight()
	consensusKey := prefixedClientKey(counterpartyConnection.ClientID, host.KeyConsensusState(consensusHeight))
	proofConsensus, _ := counterparty.QueryProof(consensusKey)

	msg := connectiontypes.NewMsgConnectionOpenTry(
		connection.ID, connection.ClientID,
		counterpartyConnection.ID, counterpartyConnection.ClientID,
		counterparty.GetPrefix(), []string{ConnectionVersion},
		proofInit, proofConsensus,
		proofHeight, consensusHeight,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ConnectionOpenAck will construct and execute a MsgConnectionOpenAck.
func (chain *TestChain) ConnectionOpenAck(
	counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
) error {
	connectionKey := host.KeyConnection(counterpartyConnection.ID)
	proofTry, proofHeight := counterparty.QueryProof(connectionKey)

	// retrieve consensus state to provide proof for
	consState, found := counterparty.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(counterparty.GetContext(), counterpartyConnection.ClientID)
	require.True(chain.t, found)

	consensusHeight := consState.GetHeight()
	consensusKey := prefixedClientKey(counterpartyConnection.ClientID, host.KeyConsensusState(consensusHeight))
	proofConsensus, _ := counterparty.QueryProof(consensusKey)

	msg := connectiontypes.NewMsgConnectionOpenAck(
		connection.ID,
		proofTry, proofConsensus,
		proofHeight, consensusHeight,
		ConnectionVersion,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ConnectionOpenConfirm will construct and execute a MsgConnectionOpenConfirm.
func (chain *TestChain) ConnectionOpenConfirm(
	counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
) error {
	connectionKey := host.KeyConnection(counterpartyConnection.ID)
	proof, height := counterparty.QueryProof(connectionKey)

	msg := connectiontypes.NewMsgConnectionOpenConfirm(
		connection.ID,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// CreatePortCapability binds and claims a capability for the given portID if it does not
// already exist. This function will fail testing on any resulting error.
func (chain *TestChain) CreatePortCapability(portID string) {
	// check if the portId is already binded, if not bind it
	_, ok := chain.App.ScopedIBCKeeper.GetCapability(chain.GetContext(), host.PortPath(portID))
	if !ok {
		cap, err := chain.App.ScopedIBCKeeper.NewCapability(chain.GetContext(), host.PortPath(portID))
		require.NoError(chain.t, err)
		err = chain.App.ScopedTransferKeeper.ClaimCapability(chain.GetContext(), cap, host.PortPath(portID))
		require.NoError(chain.t, err)
	}

	chain.App.Commit()

	chain.NextBlock()
}

// GetPortCapability returns the port capability for the given portID. The capability must
// exist, otherwise testing will fail.
func (chain *TestChain) GetPortCapability(portID string) *capabilitytypes.Capability {
	cap, ok := chain.App.ScopedIBCKeeper.GetCapability(chain.GetContext(), host.PortPath(portID))
	require.True(chain.t, ok)

	return cap
}

// CreateChannelCapability binds and claims a capability for the given portID and channelID
// if it does not already exist. This function will fail testing on any resulting error.
func (chain *TestChain) CreateChannelCapability(portID, channelID string) {
	capName := host.ChannelCapabilityPath(portID, channelID)
	// check if the portId is already binded, if not bind it
	_, ok := chain.App.ScopedIBCKeeper.GetCapability(chain.GetContext(), capName)
	if !ok {
		cap, err := chain.App.ScopedIBCKeeper.NewCapability(chain.GetContext(), capName)
		require.NoError(chain.t, err)
		err = chain.App.ScopedTransferKeeper.ClaimCapability(chain.GetContext(), cap, capName)
		require.NoError(chain.t, err)
	}

	chain.App.Commit()

	chain.NextBlock()
}

// GetChannelCapability returns the channel capability for the given portID and channelID.
// The capability must exist, otherwise testing will fail.
func (chain *TestChain) GetChannelCapability(portID, channelID string) *capabilitytypes.Capability {
	cap, ok := chain.App.ScopedIBCKeeper.GetCapability(chain.GetContext(), host.ChannelCapabilityPath(portID, channelID))
	require.True(chain.t, ok)

	return cap
}

// ChanOpenInit will construct and execute a MsgChannelOpenInit.
func (chain *TestChain) ChanOpenInit(
	ch, counterparty TestChannel,
	order channeltypes.Order,
	connectionID string,
) error {
	msg := channeltypes.NewMsgChannelOpenInit(
		ch.PortID, ch.ID,
		ChannelVersion, order, []string{connectionID},
		counterparty.PortID, counterparty.ID,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ChanOpenTry will construct and execute a MsgChannelOpenTry.
func (chain *TestChain) ChanOpenTry(
	counterparty *TestChain,
	ch, counterpartyCh TestChannel,
	order channeltypes.Order,
	connectionID string,
) error {
	proof, height := counterparty.QueryProof(host.KeyChannel(counterpartyCh.PortID, counterpartyCh.ID))

	msg := channeltypes.NewMsgChannelOpenTry(
		ch.PortID, ch.ID,
		ChannelVersion, order, []string{connectionID},
		counterpartyCh.PortID, counterpartyCh.ID,
		ChannelVersion,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ChanOpenAck will construct and execute a MsgChannelOpenAck.
func (chain *TestChain) ChanOpenAck(
	counterparty *TestChain,
	ch, counterpartyCh TestChannel,
) error {
	proof, height := counterparty.QueryProof(host.KeyChannel(counterpartyCh.PortID, counterpartyCh.ID))

	msg := channeltypes.NewMsgChannelOpenAck(
		ch.PortID, ch.ID,
		ChannelVersion,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ChanOpenConfirm will construct and execute a MsgChannelOpenConfirm.
func (chain *TestChain) ChanOpenConfirm(
	counterparty *TestChain,
	ch, counterpartyCh TestChannel,
) error {
	proof, height := counterparty.QueryProof(host.KeyChannel(counterpartyCh.PortID, counterpartyCh.ID))

	msg := channeltypes.NewMsgChannelOpenConfirm(
		ch.PortID, ch.ID,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ChanCloseInit will construct and execute a MsgChannelCloseInit.
//
// NOTE: does not work with ibc-transfer module
func (chain *TestChain) ChanCloseInit(
	counterparty *TestChain,
	channel TestChannel,
) error {
	msg := channeltypes.NewMsgChannelCloseInit(
		channel.PortID, channel.ID,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// SendPacket simulates sending a packet through the channel keeper. No message needs to be
// passed since this call is made from a module.
func (chain *TestChain) SendPacket(
	packet channelexported.PacketI,
) error {
	channelCap := chain.GetChannelCapability(packet.GetSourcePort(), packet.GetSourceChannel())

	// no need to send message, acting as a module
	err := chain.App.IBCKeeper.ChannelKeeper.SendPacket(chain.GetContext(), channelCap, packet)
	if err != nil {
		return err
	}

	// commit changes
	chain.App.Commit()
	chain.NextBlock()

	return nil
}

// PacketExecuted simulates receiving and wiritng an acknowledgement to the chain.
func (chain *TestChain) PacketExecuted(
	packet channelexported.PacketI,
) error {
	channelCap := chain.GetChannelCapability(packet.GetSourcePort(), packet.GetSourceChannel())

	// no need to send message, acting as a handler
	err := chain.App.IBCKeeper.ChannelKeeper.PacketExecuted(chain.GetContext(), channelCap, packet, TestHash)
	if err != nil {
		return err
	}

	// commit changes
	chain.App.Commit()
	chain.NextBlock()

	return nil
}
