package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmmath "github.com/tendermint/tendermint/libs/math"
	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
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

	ConnectionVersion = "1.0"
	ChannelVersion    = "1.0"

	ClientIDPrefix     = "clientFor"
	ConnectionIDPrefix = "connectionid"
	ChannelIDPrefix    = "channelid"
	PortIDPrefix       = "portid"
)

var (
	DefaultTrustLevel tmmath.Fraction = lite.DefaultTrustLevel
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
	ClientIDs   []string         // ClientID's used on this chain
	Connections []TestConnection // track connectionID's created for this chain
	Channels    []TestChannel    // track portID/channelID's created for this chain
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

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false,
		abci.Header{
			Height: 1,
			Time:   globalStartTime,
		},
	)

	// generate and set SenderAccount
	senderPrivKey := secp256k1.GenPrivKey()
	simapp.AddTestAddrsFromPubKeys(app, ctx, []crypto.PubKey{senderPrivKey.PubKey()}, sdk.NewInt(10000000000))
	acc := app.AccountKeeper.GetAccount(ctx, sdk.AccAddress(senderPrivKey.PubKey().Address()))

	// commit init chain changes so create client can be called by a counterparty chain
	app.Commit()
	// create current header and call begin block
	header := abci.Header{
		Height: 2,
		Time:   globalStartTime.Add(timeIncrement),
	}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	lastHeader := ibctmtypes.CreateTestHeader(chainID, 1, globalStartTime, valSet, signers)

	// create an account to send transactions from
	return &TestChain{
		t:             t,
		ChainID:       chainID,
		App:           app,
		LastHeader:    lastHeader,
		CurrentHeader: header,
		Querier:       keeper.NewQuerier(*app.IBCKeeper),
		Vals:          valSet,
		Signers:       signers,
		senderPrivKey: senderPrivKey,
		SenderAccount: acc,
	}
}

// GetContext returns the current context for the application.
func (chain *TestChain) GetContext() sdk.Context {
	return chain.App.BaseApp.NewContext(false, chain.CurrentHeader)
}

// QueryProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the query was performed.
func (chain *TestChain) QueryProof(key []byte) ([]byte, uint64) {
	res := chain.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
		Height: chain.App.LastBlockHeight(),
		Data:   key,
		Prove:  true,
	})

	merkleProof := commitmenttypes.MerkleProof{
		Proof: res.Proof,
	}

	proof, err := chain.App.AppCodec().MarshalBinaryBare(&merkleProof)
	require.NoError(chain.t, err)

	return proof, uint64(res.Height)
}

// NextBlock sets the last header to the current header and increments the current header to be
// at the next block height. It does not update the time as that is handled by the Coordinator.
//
// CONTRACT: this function must only be called after app.Commit() occurs
func (chain *TestChain) NextBlock() {
	// set the last header to the current header
	chain.LastHeader = ibctmtypes.CreateTestHeader(
		chain.CurrentHeader.ChainID,
		chain.CurrentHeader.Height,
		chain.CurrentHeader.Time,
		chain.Vals, chain.Signers,
	)

	// increment the current header
	chain.CurrentHeader = abci.Header{
		Height: chain.CurrentHeader.Height + 1,
		Time:   chain.CurrentHeader.Time,
	}
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

// NewClientID appends a new clientID string in the format:
// ClientFor<counterparty-chain-id><index>
func (chain *TestChain) NewClientID(counterpartyChainID string) string {
	clientID := ClientIDPrefix + counterpartyChainID + string(len(chain.ClientIDs))

	chain.ClientIDs = append(chain.ClientIDs, clientID)
	return clientID
}

// NewConnection appends a new TestConnection which contains references to the connection id,
// client id and counterparty client id. The connection id format:
// connectionid<index>
func (chain *TestChain) NewTestConnection(clientID, counterpartyClientID string) TestConnection {
	connectionID := ConnectionIDPrefix + string(len(chain.Connections))
	conn := TestConnection{
		ID:                   connectionID,
		ClientID:             clientID,
		CounterpartyClientID: counterpartyClientID,
	}

	chain.Connections = append(chain.Connections, conn)
	return conn
}

// NewTestChannel appends a new TestChannel which contains references to the port and channel ID
// used for channel creation and interaction. The channel id and port id format:
// channelid<index>
// portid<index>
func (chain *TestChain) NewTestChannel() TestChannel {
	portID := PortIDPrefix + string(len(chain.Channels))
	channelID := ChannelIDPrefix + string(len(chain.Channels))
	channel := TestChannel{
		PortID:    portID,
		ChannelID: channelID,
	}

	chain.Channels = append(chain.Channels, channel)

	return channel
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

// ConnectionOpenInit will construct and execute a MsgConnectionOpenInit.
func (chain *TestChain) ConnectionOpenInit(
	counterparty *TestChain,
	connection, counterpartyConnection TestConnection,
) error {
	prefix := commitmenttypes.NewMerklePrefix(counterparty.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes())

	msg := connectiontypes.NewMsgConnectionOpenInit(
		connection.ID, connection.ClientID,
		counterpartyConnection.ID, connection.CounterpartyClientID,
		prefix,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ConnectionOpenTry will construct and execute a MsgConnectionOpenTry.
func (chain *TestChain) ConnectionOpenTry(
	counterparty *TestChain,
	connection, counterpartyConnection TestConnection,
) error {
	prefix := commitmenttypes.NewMerklePrefix(counterparty.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes())

	connectionKey := host.KeyConnection(counterpartyConnection.ID)
	proofInit, proofHeight := counterparty.QueryProof(connectionKey)

	consensusHeight := uint64(counterparty.App.LastBlockHeight())
	consensusKey := prefixedClientKey(connection.ClientID, host.KeyConsensusState(consensusHeight))
	proofConsensus, _ := counterparty.QueryProof(consensusKey)

	msg := connectiontypes.NewMsgConnectionOpenTry(
		connection.ID, connection.ClientID,
		counterpartyConnection.ID, connection.CounterpartyClientID,
		prefix, []string{ConnectionVersion},
		proofInit, proofConsensus,
		proofHeight, consensusHeight,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ConnectionOpenAck will construct and execute a MsgConnectionOpenAck.
func (chain *TestChain) ConnectionOpenAck(
	counterparty *TestChain,
	connection, counterpartyConnection TestConnection,
) error {
	connectionKey := host.KeyConnection(counterpartyConnection.ID)
	proofTry, proofHeight := counterparty.QueryProof(connectionKey)

	consensusHeight := uint64(counterparty.App.LastBlockHeight())
	consensusKey := prefixedClientKey(connection.ClientID, host.KeyConsensusState(consensusHeight))
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
	connection, counterpartyConnection TestConnection,
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

// ChannelOpenInit will construct and execute a MsgChannelOpenInit.
func (chain *TestChain) ChannelOpenInit(
	ch, counterparty TestChannel,
	order channeltypes.Order,
	connectionID string,
) error {
	msg := channeltypes.NewMsgChannelOpenInit(
		ch.PortID, ch.ChannelID,
		ChannelVersion, order, []string{connectionID},
		counterparty.PortID, counterparty.ChannelID,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ChannelOpenTry will construct and execute a MsgChannelOpenTry.
func (chain *TestChain) ChannelOpenTry(
	ch, counterparty TestChannel,
	order channeltypes.Order,
	connectionID string,
) error {
	proof, height := chain.QueryProof(host.KeyConnection(connectionID))

	msg := channeltypes.NewMsgChannelOpenTry(
		ch.PortID, ch.ChannelID,
		ChannelVersion, order, []string{connectionID},
		counterparty.PortID, counterparty.ChannelID,
		ChannelVersion,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ChannelOpenAck will construct and execute a MsgChannelOpenAck.
func (chain *TestChain) ChannelOpenAck(
	ch, counterparty TestChannel,
	connectionID string,
) error {
	proof, height := chain.QueryProof(host.KeyConnection(connectionID))

	msg := channeltypes.NewMsgChannelOpenAck(
		ch.PortID, ch.ChannelID,
		ChannelVersion,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}

// ChannelOpenConfirm will construct and execute a MsgChannelOpenConfirm.
func (chain *TestChain) ChannelOpenConfirm(
	ch, counterparty TestChannel,
	connectionID string,
) error {
	proof, height := chain.QueryProof(host.KeyConnection(connectionID))

	msg := channeltypes.NewMsgChannelOpenConfirm(
		ch.PortID, ch.ChannelID,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.SendMsg(msg)
}
