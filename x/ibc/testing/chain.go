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
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
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

	ChannelVersion = "1.0"
)

// TODO: rework timing of both chains, updating clients must be at least a second difference.

// TestChain is a testing struct that wraps a simapp with the last TM Header, the current ABCI
// header and the validators of the TestChain.It also contains a field called ClientID. This
// is the clientID that *other* chains use to refer to this TestChain. For simplicity's sake
// it is also the chainID on the TestChain Header. The senderAccount is used for delivering
// transactions through the application state.
type TestChain struct {
	t *testing.T

	ClientID      string
	App           *simapp.SimApp
	LastHeader    ibctmtypes.Header // header for last block height committed
	CurrentHeader abci.Header       // header for current block height
	Querier       sdk.Querier

	Vals    *tmtypes.ValidatorSet
	Signers []tmtypes.PrivValidator

	senderPrivKey crypto.PrivKey
	senderAccount authtypes.AccountI
}

// NewTestChain initializes a new TestChain instance with a single validator set using a
// generated private key. It also creates a sender account to be used for delivering transactions.
//
// The first block height is committed to state in order to allow for client creations on counterparty
// chains. The TestChain will return with a block height starting at 2.
//
// For each block, time.Now() is used to prevent counterparty chains from falling behind,
// otherwise client updates may appear to come from the future.
func NewTestChain(t *testing.T, clientID string) *TestChain {
	// generate validator private/public key
	privVal := tmtypes.NewMockPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := []tmtypes.PrivValidator{privVal}

	startTime := time.Now()

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false,
		abci.Header{
			Height: 1,
			Time:   startTime,
		},
	)

	// generate and set senderAccount
	senderPrivKey := secp256k1.GenPrivKey()
	simapp.AddTestAddrsFromPubKeys(app, ctx, []crypto.PubKey{senderPrivKey.PubKey()}, sdk.NewInt(10000000000))
	acc := app.AccountKeeper.GetAccount(ctx, sdk.AccAddress(senderPrivKey.PubKey().Address()))

	// commit init chain changes so create client can be called by a counterparty chain
	app.Commit()
	// create current header and call begin block
	header := abci.Header{
		Height: 2,
		Time:   time.Now(),
	}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	lastHeader := ibctmtypes.CreateTestHeader(clientID, 1, startTime, valSet, signers)

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
		senderAccount: acc,
	}
}

// GetContext returns the current context for the application.
func (chain *TestChain) GetContext() sdk.Context {
	return chain.App.BaseApp.NewContext(false, chain.CurrentHeader)
}

// CommitNBlocks commits n blocks to state and updates the block height by 1 for each commit.
func (chain *TestChain) CommitNBlocks(n int) {
	for i := 0; i < n; i++ {
		chain.App.Commit()
		nextBlock(chain)
	}

}

// CommitBlockWithNewTimestamp commits the current block and starts the next block with the
// provided timestamp.
func (chain *TestChain) CommitBlockWithNewTimestamp(timestamp int64) {
	chain.App.Commit()
	nextBlock(chain)
	chain.App.BeginBlock(abci.RequestBeginBlock{
		Header: abci.Header{
			Height: chain.CurrentHeader.Height,
			Time:   time.Unix(timestamp, 0),
		}},
	)
}

// nextBlock sets the last header to the current header and increments the current header to be
// at the next block height and time.
//
// CONTRACT: this function must only be called after app.Commit() occurs
func nextBlock(chain *TestChain) {
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
		Time:   time.Now(),
	}

	chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})
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
		[]uint64{chain.senderAccount.GetAccountNumber()},
		[]uint64{chain.senderAccount.GetSequence()},
		true, true, chain.senderPrivKey,
	)
	require.NoError(chain.t, err)

	// SignCheckDeliver calls app.Commit()
	nextBlock(chain)

	// increment sequence for successful transaction execution
	chain.senderAccount.SetSequence(chain.senderAccount.GetSequence() + 1)
}

// CreateClient will construct and execute a 07-tendermint MsgCreateClient. A counterparty
// client will be created on the (target) chain.
func (chain *TestChain) CreateClient(counterparty *TestChain) {
	// commit counterparty's state
	counterparty.CommitNBlocks(1)
	chain.CommitNBlocks(1)

	// construct MsgCreateClient using counterparty
	msg := ibctmtypes.NewMsgCreateClient(
		counterparty.ClientID, counterparty.LastHeader,
		lite.DefaultTrustLevel, TrustingPeriod, UnbondingPeriod, MaxClockDrift,
		chain.senderAccount.GetAddress(),
	)

	chain.SendMsg(msg)
}

// UpdateClient will construct and execute a 07-tendermint MsgUpdateClient. The counterparty
// client will be updated on the (target) chain.
func (chain *TestChain) UpdateClient(counterparty *TestChain) {
	// commit counterparty's state
	counterparty.CommitNBlocks(1)
	chain.CommitNBlocks(1)

	msg := ibctmtypes.NewMsgUpdateClient(
		counterparty.ClientID, counterparty.LastHeader,
		chain.senderAccount.GetAddress(),
	)

	chain.SendMsg(msg)
}

// TODO: update with msg passing
// CreateConnection creates a connection to the counterparty with the provided state.
func (chain *TestChain) CreateConnection(
	connID, counterpartyConnID, clientID, counterpartyClientID string,
	state connectiontypes.State,
) connectiontypes.ConnectionEnd {

	chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})

	counterparty := connectiontypes.NewCounterparty(
		counterpartyClientID, counterpartyConnID,
		commitmenttypes.NewMerklePrefix(chain.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()),
	)

	connection := connectiontypes.ConnectionEnd{
		State:        state,
		ID:           connID,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     connectiontypes.GetCompatibleVersions(),
	}

	ctx := chain.GetContext()
	chain.App.IBCKeeper.ConnectionKeeper.SetConnection(ctx, connID, connection)
	return connection
}

// TODO: update with msg passing
// CreateChannel constructs a channel with the given counterparty and connectionID.
func (chain *TestChain) CreateChannel(
	portID, channelID, counterpartyPortID, counterpartyChannelID string,
	state channeltypes.State, order channeltypes.Order, connectionID string,
) channeltypes.Channel {

	counterparty := channeltypes.NewCounterparty(counterpartyPortID, counterpartyChannelID)

	// sets channel with given state
	channel := channeltypes.NewChannel(state, order, counterparty,
		[]string{connectionID}, ChannelVersion,
	)

	ctx := chain.GetContext()
	chain.App.IBCKeeper.ChannelKeeper.SetChannel(ctx, portID, channelID, channel)
	chain.NewCapability(portID, channelID)
	return channel
}

// NewCapability creates a new capability for the provided port and channel ID
// if the capability does not already exist.
func (chain *TestChain) NewCapability(portID, channelID string) (*capabilitytypes.Capability, error) {
	ctx := chain.GetContext()

	capName := host.ChannelCapabilityPath(portID, channelID)
	cap, ok := chain.App.ScopedIBCKeeper.GetCapability(ctx, capName)
	if !ok {
		return chain.App.ScopedIBCKeeper.NewCapability(ctx, capName)
	}

	return cap, nil
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
