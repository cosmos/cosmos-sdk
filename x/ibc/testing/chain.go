package testing

import (
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10

	ChannelVersion = "1.0"
)

// TestChain is a testing struct that wraps a simapp with the latest Header, Vals and Signers.
// It also contains a field called ClientID. This is the clientID that *other* chains use
// to refer to this TestChain. For simplicity's sake it is also the chainID on the TestChain
// Header.
type TestChain struct {
	ClientID string
	App      *simapp.SimApp
	Header   ibctmtypes.Header
	Vals     *tmtypes.ValidatorSet
	Signers  []tmtypes.PrivValidator
	Querier  sdk.Querier
}

// NewTestChain initializes a new TestChain instance with a single validator set using a
// generated private key.
func NewTestChain(clientID string) *TestChain {
	privVal := tmtypes.NewMockPV()

	pubKey, err := privVal.GetPubKey()
	if err != nil {
		panic(err)
	}

	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signers := []tmtypes.PrivValidator{privVal}
	now := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	header := ibctmtypes.CreateTestHeader(clientID, 1, now, valSet, signers)

	app := simapp.Setup(false)

	return &TestChain{
		ClientID: clientID,
		App:      app,
		Header:   header,
		Vals:     valSet,
		Signers:  signers,
		Querier:  keeper.NewQuerier(*app.IBCKeeper),
	}
}

// GetContext creates a simple context for testing purposes.
func (chain *TestChain) GetContext() sdk.Context {
	return chain.App.BaseApp.NewContext(false, abci.Header{ChainID: chain.Header.SignedHeader.Header.ChainID, Height: int64(chain.Header.GetHeight())})
}

// CreateClient will create a client on the chain using the provided counterparty
// TestChain.
func (chain *TestChain) CreateClient(counterparty *TestChain) error {
	counterparty.Header = nextHeader(counterparty)

	// commit and create a new block on the counterparty chain to get a fresh CommitID
	counterparty.App.Commit()
	commitID := counterparty.App.LastCommitID()
	counterparty.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: int64(counterparty.Header.GetHeight()), Time: counterparty.Header.Time}})

	// set HistoricalInfo on counterparty chain after Commit
	ctxCounterparty := counterparty.GetContext()
	validator := stakingtypes.NewValidator(
		sdk.ValAddress(counterparty.Vals.Validators[0].Address), counterparty.Vals.Validators[0].PubKey, stakingtypes.Description{},
	)
	validator.Status = sdk.Bonded
	validator.Tokens = sdk.NewInt(1000000) // get one voting power
	validators := []stakingtypes.Validator{validator}
	histInfo := stakingtypes.HistoricalInfo{
		Header: abci.Header{
			AppHash: commitID.Hash,
		},
		Valset: validators,
	}
	counterparty.App.StakingKeeper.SetHistoricalInfo(ctxCounterparty, int64(counterparty.Header.GetHeight()), histInfo)

	// create target ctx
	ctxTarget := chain.GetContext()

	// create client
	clientState, err := ibctmtypes.Initialize(counterparty.ClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, counterparty.Header)
	if err != nil {
		return err
	}
	_, err = chain.App.IBCKeeper.ClientKeeper.CreateClient(ctxTarget, clientState, counterparty.Header.ConsensusState())
	if err != nil {
		return err
	}
	return nil

	// TODO: use simapp SignCheckDeliver to update state
	// _, _, err := simapp.SignCheckDeliver(
	// 	suite.T(),
	// 	suite.cdc,
	// 	suite.app.BaseApp,
	// 	ctx.BlockHeader(),
	// 	[]sdk.Msg{clienttypes.NewMsgCreateClient(clientID, clientexported.ClientTypeTendermint, consState, accountAddress)},
	// 	[]uint64{baseAccount.GetAccountNumber()},
	// 	[]uint64{baseAccount.GetSequence()},
	// 	true, true, accountPrivKey,
	// )
}

// UpdateClient will update a client on the chain using the provided counterparty
// TestChain.
func (chain *TestChain) UpdateClient(counterparty *TestChain) {
	// Create target ctx
	ctxTarget := chain.GetContext()

	// if clientState does not already exist, return without updating
	_, found := chain.App.IBCKeeper.ClientKeeper.GetClientState(
		ctxTarget, counterparty.ClientID,
	)
	if !found {
		return // TODO: change to panic
	}

	// commit and begin a new block when updating a client
	counterparty.App.Commit()
	commitID := counterparty.App.LastCommitID()

	counterparty.Header = nextHeader(counterparty)
	counterparty.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: int64(counterparty.Header.GetHeight()), Time: counterparty.Header.Time}})

	// Set HistoricalInfo on counterparty chain after Commit
	ctxCounterparty := counterparty.GetContext()
	validator := stakingtypes.NewValidator(
		sdk.ValAddress(counterparty.Vals.Validators[0].Address), counterparty.Vals.Validators[0].PubKey, stakingtypes.Description{},
	)
	validator.Status = sdk.Bonded
	validator.Tokens = sdk.NewInt(1000000)
	validators := []stakingtypes.Validator{validator}
	histInfo := stakingtypes.HistoricalInfo{
		Header: abci.Header{
			AppHash: commitID.Hash,
		},
		Valset: validators,
	}
	counterparty.App.StakingKeeper.SetHistoricalInfo(ctxCounterparty, int64(counterparty.Header.GetHeight()), histInfo)

	consensusState := ibctmtypes.ConsensusState{
		Height:       counterparty.Header.GetHeight(),
		Timestamp:    counterparty.Header.Time,
		Root:         commitmenttypes.NewMerkleRoot(commitID.Hash),
		ValidatorSet: counterparty.Vals,
	}

	chain.App.IBCKeeper.ClientKeeper.SetClientConsensusState(
		ctxTarget, counterparty.ClientID, counterparty.Header.GetHeight(), consensusState,
	)
	chain.App.IBCKeeper.ClientKeeper.SetClientState(
		ctxTarget, ibctmtypes.NewClientState(counterparty.ClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, counterparty.Header),
	)

	// TODO: use simapp SignCheckDeliver to update state
	// _, _, err := simapp.SignCheckDeliver(
	// 	suite.T(),
	// 	suite.cdc,
	// 	suite.app.BaseApp,
	// 	ctx.BlockHeader(),
	// 	[]sdk.Msg{clienttypes.NewMsgUpdateClient(clientID, suite.header, accountAddress)},
	// 	[]uint64{baseAccount.GetAccountNumber()},
	// 	[]uint64{baseAccount.GetSequence()},
	// 	true, true, accountPrivKey,
	// )
	// suite.Require().NoError(err)
}

// CreateConnection creates a connection to the counterparty with the provided state.
func (chain *TestChain) CreateConnection(
	connID, counterpartyConnID, clientID, counterpartyClientID string,
	state connectiontypes.State,
) connectiontypes.ConnectionEnd {

	counterparty := connectiontypes.NewCounterparty(counterpartyClientID, counterpartyConnID, commitmenttypes.NewMerklePrefix(chain.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))

	connection := connectiontypes.ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     connectiontypes.GetCompatibleVersions(),
	}

	ctx := chain.GetContext()
	chain.App.IBCKeeper.ConnectionKeeper.SetConnection(ctx, connID, connection)
	return connection
}

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
	return channel
}

// nextHeader increments the header height by 1 and increments the timestamp by 1 minute.
func nextHeader(chain *TestChain) ibctmtypes.Header {
	return ibctmtypes.CreateTestHeader(chain.Header.SignedHeader.Header.ChainID, int64(chain.Header.GetHeight())+1,
		chain.Header.Time.Add(time.Minute), chain.Vals, chain.Signers)
}

// CommitNBlocks commits n blocks to state and updates the block height by 1 for each commit.
func (chain *TestChain) CommitNBlocks(n int) {
	for i := 0; i < n; i++ {
		chain.App.Commit()
		chain.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: chain.App.LastBlockHeight() + 1}})
	}
}

// CommitBlockWithNewTimestamp commits the current block and starts the next block with the provided timestamp.
func (chain *TestChain) CommitBlockWithNewTimestamp(timestamp int64) {
	chain.App.Commit()
	chain.App.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: chain.App.LastBlockHeight() + 1, Time: time.Unix(timestamp, 0)}})
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
