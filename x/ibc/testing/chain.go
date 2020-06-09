package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
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

	ConnectionIDPrefix = "connectionID"
	ChannelIDPrefix    = "channelID"
	PortIDPrefix       = "portID"
)

// TestChain is a testing struct that wraps a simapp with the last TM Header, the current ABCI
// header and the validators of the TestChain. It also contains a field called ClientID. This
// is the clientID that *other* chains use to refer to this TestChain. For simplicity's sake
// it is also the chainID on the TestChain Header. The SenderAccount is used for delivering
// transactions through the application state.
type TestChain struct {
	t *testing.T

	App           *simapp.SimApp
	LastHeader    ibctmtypes.Header // header for last block height committed
	CurrentHeader abci.Header       // header for current block height
	Querier       sdk.Querier

	Vals    *tmtypes.ValidatorSet
	Signers []tmtypes.PrivValidator

	senderPrivKey crypto.PrivKey
	SenderAccount authtypes.AccountI

	// IBC specific helpers
	ClientID      string
	Connections []string  // track connectionID's created for this chain
	Channels    []Channel // track portID/channelID's created for this chain
}

// NewTestChain initializes a new TestChain instance with a single validator set using a
// generated private key. It also creates a sender account to be used for delivering transactions.
//
// The first block height is committed to state in order to allow for client creations on counterparty
// chains. The TestChain will return with a block height starting at 2.
//
// Time management is handled by the IBCTestSuite in order to ensure synchrony between chains. Each
// update of any chain increments the block header time by 5 seconds.
func NewTestChain(t *testing.T, clientID string) *TestChain {
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

	lastHeader := ibctmtypes.CreateTestHeader(clientID, 1, globalStartTime, valSet, signers)

	// create an account to send transactions from
	return &TestChain{
		t:             t,
		ClientID:      clientID,
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

// QueryProof performs an abci query with the given key and returns the merkle proof for the query
// and the height at which the query was performed.
func (chain *TestChain) QueryProof(key []byte) (commitmenttypes.MerkleProof, uint64) {
	res := chain.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
		Height: chain.App.LastBlockHeight(),
		Data:   key,
		Prove:  true,
	})

	proof := commitmenttypes.MerkleProof{
		Proof: res.Proof,
	}

	return proof, uint64(res.Height)
}

// nextBlock sets the last header to the current header and increments the current header to be
// at the next block height. It does not update the time as that is handled by the IBCTestSuite.
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
func (chain *TestChain) SendMsg(msg sdk.Msg) {
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
	require.NoError(chain.t, err)

	// SignCheckDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
}

// NewConnection appends a new connectionID string in the format:
// connectionID<index>
func (chain *TestChain) NewConnection() string {
	conn := ConnectionIDPrefix + string(len(chain.Connections))

	chain.Connections = append(chain.Connections, conn)
	return conn
}

// NewChannel appends a new testing channel which contains references to the port and channel ID
// used for channel creation and interaction.
func (chain *TestChain) NewChannel() Channel {
	portID := PortIDPrefix + string(len(chain.Channels))
	channelID := ChannelIDPrefix + string(len(chain.Channels))
	channel := NewChannel(portID, channelID)

	chain.Channels = append(chain.Channels, channel)
	return channel
}

// CreateClient will construct and execute a 07-tendermint MsgCreateClient. A counterparty
// client will be created on the (target) chain.
func (chain *TestChain) CreateClient(counterparty *TestChain) {
	// construct MsgCreateClient using counterparty
	msg := ibctmtypes.NewMsgCreateClient(
		counterparty.ClientID, counterparty.LastHeader,
		lite.DefaultTrustLevel, TrustingPeriod, UnbondingPeriod, MaxClockDrift,
		chain.SenderAccount.GetAddress(),
	)

	chain.SendMsg(msg)
}

// UpdateClient will construct and execute a 07-tendermint MsgUpdateClient. The counterparty
// client will be updated on the (target) chain.
func (chain *TestChain) UpdateClient(counterparty *TestChain) {
	msg := ibctmtypes.NewMsgUpdateClient(
		counterparty.ClientID, counterparty.LastHeader,
		chain.SenderAccount.GetAddress(),
	)

	chain.SendMsg(msg)
}

// ConnectionOpenInit will construct and execute a MsgConnectionOpenInit.
func (chain *TestChain) ConnectionOpenInit(
	counterparty *TestChain,
	connectionID, counterpartyConnectionID string,
) {
	prefix := commitmenttypes.NewMerklePrefix(counterparty.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes())

	msg := connectiontypes.NewMsgConnectionOpenInit(
		connectionID, chain.ClientID,
		counterpartyConnectionID, counterparty.ClientID,
		prefix,
		chain.SenderAccount.GetAddress(),
	)
	chain.SendMsg(msg)
}

// ConnectionOpenTry will construct and execute a MsgConnectionOpenTry.
func (chain *TestChain) ConnectionOpenTry(
	counterparty *TestChain,
	connectionID, counterpartyConnectionID string,
) {
	prefix := commitmenttypes.NewMerklePrefix(counterparty.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes())

	connectionKey := host.KeyConnection(counterpartyConnectionID)
	proofInit, proofHeight := counterparty.QueryProof(connectionKey)

	consensusHeight := uint64(counterparty.App.LastBlockHeight())
	consensusKey := prefixedClientKey(chain.ClientID, host.KeyConsensusState(consensusHeight))
	proofConsensus, _ := counterparty.QueryProof(consensusKey)

	msg := connectiontypes.NewMsgConnectionOpenTry(
		connectionID, chain.ClientID,
		counterpartyConnectionID, counterparty.ClientID,
		prefix, []string{ConnectionVersion},
		proofInit, proofConsensus,
		proofHeight, consensusHeight,
		chain.SenderAccount.GetAddress(),
	)
	chain.SendMsg(msg)
}

// ConnectionOpenAck will construct and execute a MsgConnectionOpenAck.
func (chain *TestChain) ConnectionOpenAck(
	counterparty *TestChain,
	connectionID, counterpartyConnectionID string,
) {
	connectionKey := host.KeyConnection(counterpartyConnectionID)
	proofTry, proofHeight := counterparty.QueryProof(connectionKey)

	consensusHeight := uint64(counterparty.App.LastBlockHeight())
	consensusKey := prefixedClientKey(chain.ClientID, host.KeyConsensusState(consensusHeight))
	proofConsensus, _ := counterparty.QueryProof(consensusKey)

	msg := connectiontypes.NewMsgConnectionOpenAck(
		connectionID,
		proofTry, proofConsensus,
		proofHeight, consensusHeight,
		ConnectionVersion,
		chain.SenderAccount.GetAddress(),
	)
	chain.SendMsg(msg)
}

// ConnectionOpenConfirm will construct and execute a MsgConnectionOpenConfirm.
func (chain *TestChain) ConnectionOpenConfirm(
	counterparty *TestChain,
	connectionID, counterpartyConnectionID string,
) {
	connectionKey := host.KeyConnection(counterpartyConnectionID)
	proof, height := counterparty.QueryProof(connectionKey)

	msg := connectiontypes.NewMsgConnectionOpenConfirm(
		connectionID,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	chain.SendMsg(msg)
}

// ChannelOpenInit will construct and execute a MsgChannelOpenInit.
func (chain *TestChain) ChannelOpenInit(
	ch, counterparty Channel,
	order channeltypes.Order,
	connectionID string,
) {
	msg := channeltypes.NewMsgChannelOpenInit(
		ch.GetPortID(), ch.GetChannelID(),
		ChannelVersion, order, []string{connectionID},
		counterparty.GetPortID(), counterparty.GetChannelID(),
		chain.SenderAccount.GetAddress(),
	)
	chain.SendMsg(msg)
}

// ChannelOpenTry will construct and execute a MsgChannelOpenTry.
func (chain *TestChain) ChannelOpenTry(
	ch, counterparty Channel,
	order channeltypes.Order,
	connectionID string,
) {
	proof, height := chain.QueryProof(host.KeyConnection(connectionID))

	msg := channeltypes.NewMsgChannelOpenTry(
		ch.GetPortID(), ch.GetChannelID(),
		ChannelVersion, order, []string{connectionID},
		counterparty.GetPortID(), counterparty.GetChannelID(),
		ChannelVersion,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	chain.SendMsg(msg)
}

// ChannelOpenAck will construct and execute a MsgChannelOpenAck.
func (chain *TestChain) ChannelOpenAck(
	ch, counterparty Channel,
	connectionID string,
) {
	proof, height := chain.QueryProof(host.KeyConnection(connectionID))

	msg := channeltypes.NewMsgChannelOpenAck(
		ch.GetPortID(), ch.GetChannelID(),
		ChannelVersion,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	chain.SendMsg(msg)
}

// ChannelOpenConfirm will construct and execute a MsgChannelOpenConfirm.
func (chain *TestChain) ChannelOpenConfirm(
	ch, counterparty Channel,
	connectionID string,
) {
	proof, height := chain.QueryProof(host.KeyConnection(connectionID))

	msg := channeltypes.NewMsgChannelOpenConfirm(
		ch.GetPortID(), ch.GetChannelID(),
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	chain.SendMsg(msg)
}
