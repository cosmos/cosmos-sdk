package ibctesting

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmprotoversion "github.com/tendermint/tendermint/proto/tendermint/version"
	tmtypes "github.com/tendermint/tendermint/types"
	tmversion "github.com/tendermint/tendermint/version"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/testing/mock"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	// Default params constants used to create a TM client
	TrustingPeriod     time.Duration = time.Hour * 24 * 7 * 2
	UnbondingPeriod    time.Duration = time.Hour * 24 * 7 * 3
	MaxClockDrift      time.Duration = time.Second * 10
	DefaultDelayPeriod uint64        = 0

	DefaultChannelVersion = ibctransfertypes.Version
	InvalidID             = "IDisInvalid"

	ConnectionIDPrefix = "conn"
	ChannelIDPrefix    = "chan"

	TransferPort = ibctransfertypes.ModuleName
	MockPort     = mock.ModuleName

	// used for testing UpdateClientProposal
	Title       = "title"
	Description = "description"
)

var (
	DefaultOpenInitVersion *connectiontypes.Version

	// Default params variables used to create a TM client
	DefaultTrustLevel ibctmtypes.Fraction = ibctmtypes.DefaultTrustLevel
	TestHash                              = tmhash.Sum([]byte("TESTING HASH"))
	TestCoin                              = sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))

	UpgradePath = []string{"upgrade", "upgradedIBCState"}

	ConnectionVersion = connectiontypes.ExportedVersionsToProto(connectiontypes.GetCompatibleVersions())[0]

	MockAcknowledgement = mock.MockAcknowledgement
	MockCommitment      = mock.MockCommitment
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
	LastHeader    *ibctmtypes.Header // header for last block height committed
	CurrentHeader tmproto.Header     // header for current block height
	QueryServer   types.QueryServer
	TxConfig      client.TxConfig
	Codec         codec.BinaryMarshaler

	Vals    *tmtypes.ValidatorSet
	Signers []tmtypes.PrivValidator

	senderPrivKey cryptotypes.PrivKey
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
	privVal := mock.NewPV()
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
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
	}

	app := simapp.SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, balance)

	// create current header and call begin block
	header := tmproto.Header{
		ChainID: chainID,
		Height:  1,
		Time:    globalStartTime,
	}

	txConfig := simapp.MakeTestEncodingConfig().TxConfig

	// create an account to send transactions from
	chain := &TestChain{
		t:             t,
		ChainID:       chainID,
		App:           app,
		CurrentHeader: header,
		QueryServer:   app.IBCKeeper,
		TxConfig:      txConfig,
		Codec:         app.AppCodec(),
		Vals:          valSet,
		Signers:       signers,
		senderPrivKey: senderPrivKey,
		SenderAccount: acc,
		ClientIDs:     make([]string, 0),
		Connections:   make([]*TestConnection, 0),
	}

	cap := chain.App.IBCKeeper.PortKeeper.BindPort(chain.GetContext(), MockPort)
	err = chain.App.ScopedIBCMockKeeper.ClaimCapability(chain.GetContext(), cap, host.PortPath(MockPort))
	require.NoError(t, err)

	chain.NextBlock()

	return chain
}

// GetContext returns the current context for the application.
func (chain *TestChain) GetContext() sdk.Context {
	return chain.App.BaseApp.NewContext(false, chain.CurrentHeader)
}

// QueryProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryProof(key []byte) ([]byte, clienttypes.Height) {
	res := chain.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
		Height: chain.App.LastBlockHeight() - 1,
		Data:   key,
		Prove:  true,
	})

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	require.NoError(chain.t, err)

	proof, err := chain.App.AppCodec().MarshalBinaryBare(&merkleProof)
	require.NoError(chain.t, err)

	revision := clienttypes.ParseChainID(chain.ChainID)

	// proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree. Thus we return proof height + 1
	return proof, clienttypes.NewHeight(revision, uint64(res.Height)+1)
}

// QueryUpgradeProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryUpgradeProof(key []byte, height uint64) ([]byte, clienttypes.Height) {
	res := chain.App.Query(abci.RequestQuery{
		Path:   "store/upgrade/key",
		Height: int64(height - 1),
		Data:   key,
		Prove:  true,
	})

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	require.NoError(chain.t, err)

	proof, err := chain.App.AppCodec().MarshalBinaryBare(&merkleProof)
	require.NoError(chain.t, err)

	revision := clienttypes.ParseChainID(chain.ChainID)

	// proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree. Thus we return proof height + 1
	return proof, clienttypes.NewHeight(revision, uint64(res.Height+1))
}

// QueryClientStateProof performs and abci query for a client state
// stored with a given clientID and returns the ClientState along with the proof
func (chain *TestChain) QueryClientStateProof(clientID string) (exported.ClientState, []byte) {
	// retrieve client state to provide proof for
	clientState, found := chain.App.IBCKeeper.ClientKeeper.GetClientState(chain.GetContext(), clientID)
	require.True(chain.t, found)

	clientKey := host.FullClientStateKey(clientID)
	proofClient, _ := chain.QueryProof(clientKey)

	return clientState, proofClient
}

// QueryConsensusStateProof performs an abci query for a consensus state
// stored on the given clientID. The proof and consensusHeight are returned.
func (chain *TestChain) QueryConsensusStateProof(clientID string) ([]byte, clienttypes.Height) {
	clientState := chain.GetClientState(clientID)

	consensusHeight := clientState.GetLatestHeight().(clienttypes.Height)
	consensusKey := host.FullConsensusStateKey(clientID, consensusHeight)
	proofConsensus, _ := chain.QueryProof(consensusKey)

	return proofConsensus, consensusHeight
}

// NextBlock sets the last header to the current header and increments the current header to be
// at the next block height. It does not update the time as that is handled by the Coordinator.
//
// CONTRACT: this function must only be called after app.Commit() occurs
func (chain *TestChain) NextBlock() {
	// set the last header to the current header
	// use nil trusted fields
	chain.LastHeader = chain.CurrentTMClientHeader()

	// increment the current header
	chain.CurrentHeader = tmproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.CurrentHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.Vals.Hash(),
	}

	chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})

}

// sendMsgs delivers a transaction through the application without returning the result.
func (chain *TestChain) sendMsgs(msgs ...sdk.Msg) error {
	_, err := chain.SendMsgs(msgs...)
	return err
}

// SendMsgs delivers a transaction through the application. It updates the senders sequence
// number and updates the TestChain's headers. It returns the result and error if one
// occurred.
func (chain *TestChain) SendMsgs(msgs ...sdk.Msg) (*sdk.Result, error) {
	_, r, err := simapp.SignCheckDeliver(
		chain.t,
		chain.TxConfig,
		chain.App.BaseApp,
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		true, true, chain.senderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	// SignCheckDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)

	return r, nil
}

// GetClientState retrieves the client state for the provided clientID. The client is
// expected to exist otherwise testing will fail.
func (chain *TestChain) GetClientState(clientID string) exported.ClientState {
	clientState, found := chain.App.IBCKeeper.ClientKeeper.GetClientState(chain.GetContext(), clientID)
	require.True(chain.t, found)

	return clientState
}

// GetConsensusState retrieves the consensus state for the provided clientID and height.
// It will return a success boolean depending on if consensus state exists or not.
func (chain *TestChain) GetConsensusState(clientID string, height exported.Height) (exported.ConsensusState, bool) {
	return chain.App.IBCKeeper.ClientKeeper.GetClientConsensusState(chain.GetContext(), clientID, height)
}

// GetValsAtHeight will return the validator set of the chain at a given height. It will return
// a success boolean depending on if the validator set exists or not at that height.
func (chain *TestChain) GetValsAtHeight(height int64) (*tmtypes.ValidatorSet, bool) {
	histInfo, ok := chain.App.StakingKeeper.GetHistoricalInfo(chain.GetContext(), height)
	if !ok {
		return nil, false
	}

	valSet := stakingtypes.Validators(histInfo.Valset)

	tmValidators, err := teststaking.ToTmValidators(valSet)
	if err != nil {
		panic(err)
	}
	return tmtypes.NewValidatorSet(tmValidators), true
}

// GetConnection retrieves an IBC Connection for the provided TestConnection. The
// connection is expected to exist otherwise testing will fail.
func (chain *TestChain) GetConnection(testConnection *TestConnection) connectiontypes.ConnectionEnd {
	connection, found := chain.App.IBCKeeper.ConnectionKeeper.GetConnection(chain.GetContext(), testConnection.ID)
	require.True(chain.t, found)

	return connection
}

// GetChannel retrieves an IBC Channel for the provided TestChannel. The channel
// is expected to exist otherwise testing will fail.
func (chain *TestChain) GetChannel(testChannel TestChannel) channeltypes.Channel {
	channel, found := chain.App.IBCKeeper.ChannelKeeper.GetChannel(chain.GetContext(), testChannel.PortID, testChannel.ID)
	require.True(chain.t, found)

	return channel
}

// GetAcknowledgement retrieves an acknowledgement for the provided packet. If the
// acknowledgement does not exist then testing will fail.
func (chain *TestChain) GetAcknowledgement(packet exported.PacketI) []byte {
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
func (chain *TestChain) NewClientID(clientType string) string {
	clientID := fmt.Sprintf("%s-%s", clientType, strconv.Itoa(len(chain.ClientIDs)))
	chain.ClientIDs = append(chain.ClientIDs, clientID)
	return clientID
}

// AddTestConnection appends a new TestConnection which contains references
// to the connection id, client id and counterparty client id.
func (chain *TestChain) AddTestConnection(clientID, counterpartyClientID string) *TestConnection {
	conn := chain.ConstructNextTestConnection(clientID, counterpartyClientID)

	chain.Connections = append(chain.Connections, conn)
	return conn
}

// ConstructNextTestConnection constructs the next test connection to be
// created given a clientID and counterparty clientID. The connection id
// format: <chainID>-conn<index>
func (chain *TestChain) ConstructNextTestConnection(clientID, counterpartyClientID string) *TestConnection {
	connectionID := connectiontypes.FormatConnectionIdentifier(uint64(len(chain.Connections)))
	return &TestConnection{
		ID:                   connectionID,
		ClientID:             clientID,
		NextChannelVersion:   DefaultChannelVersion,
		CounterpartyClientID: counterpartyClientID,
	}
}

// GetFirstTestConnection returns the first test connection for a given clientID.
// The connection may or may not exist in the chain state.
func (chain *TestChain) GetFirstTestConnection(clientID, counterpartyClientID string) *TestConnection {
	if len(chain.Connections) > 0 {
		return chain.Connections[0]
	}

	return chain.ConstructNextTestConnection(clientID, counterpartyClientID)
}

// AddTestChannel appends a new TestChannel which contains references to the port and channel ID
// used for channel creation and interaction. See 'NextTestChannel' for channel ID naming format.
func (chain *TestChain) AddTestChannel(conn *TestConnection, portID string) TestChannel {
	channel := chain.NextTestChannel(conn, portID)
	conn.Channels = append(conn.Channels, channel)
	return channel
}

// NextTestChannel returns the next test channel to be created on this connection, but does not
// add it to the list of created channels. This function is expected to be used when the caller
// has not created the associated channel in app state, but would still like to refer to the
// non-existent channel usually to test for its non-existence.
//
// channel ID format: <connectionid>-chan<channel-index>
//
// The port is passed in by the caller.
func (chain *TestChain) NextTestChannel(conn *TestConnection, portID string) TestChannel {
	nextChanSeq := chain.App.IBCKeeper.ChannelKeeper.GetNextChannelSequence(chain.GetContext())
	channelID := channeltypes.FormatChannelIdentifier(nextChanSeq)
	return TestChannel{
		PortID:               portID,
		ID:                   channelID,
		ClientID:             conn.ClientID,
		CounterpartyClientID: conn.CounterpartyClientID,
		Version:              conn.NextChannelVersion,
	}
}

// ConstructMsgCreateClient constructs a message to create a new client state (tendermint or solomachine).
// NOTE: a solo machine client will be created with an empty diversifier.
func (chain *TestChain) ConstructMsgCreateClient(counterparty *TestChain, clientID string, clientType string) *clienttypes.MsgCreateClient {
	var (
		clientState    exported.ClientState
		consensusState exported.ConsensusState
	)

	switch clientType {
	case exported.Tendermint:
		height := counterparty.LastHeader.GetHeight().(clienttypes.Height)
		clientState = ibctmtypes.NewClientState(
			counterparty.ChainID, DefaultTrustLevel, TrustingPeriod, UnbondingPeriod, MaxClockDrift,
			height, commitmenttypes.GetSDKSpecs(), UpgradePath, false, false,
		)
		consensusState = counterparty.LastHeader.ConsensusState()
	case exported.Solomachine:
		solo := NewSolomachine(chain.t, chain.Codec, clientID, "", 1)
		clientState = solo.ClientState()
		consensusState = solo.ConsensusState()
	default:
		chain.t.Fatalf("unsupported client state type %s", clientType)
	}

	msg, err := clienttypes.NewMsgCreateClient(
		clientState, consensusState, chain.SenderAccount.GetAddress(),
	)
	require.NoError(chain.t, err)
	return msg
}

// CreateTMClient will construct and execute a 07-tendermint MsgCreateClient. A counterparty
// client will be created on the (target) chain.
func (chain *TestChain) CreateTMClient(counterparty *TestChain, clientID string) error {
	// construct MsgCreateClient using counterparty
	msg := chain.ConstructMsgCreateClient(counterparty, clientID, exported.Tendermint)
	return chain.sendMsgs(msg)
}

// UpdateTMClient will construct and execute a 07-tendermint MsgUpdateClient. The counterparty
// client will be updated on the (target) chain. UpdateTMClient mocks the relayer flow
// necessary for updating a Tendermint client.
func (chain *TestChain) UpdateTMClient(counterparty *TestChain, clientID string) error {
	header, err := chain.ConstructUpdateTMClientHeader(counterparty, clientID)
	require.NoError(chain.t, err)

	msg, err := clienttypes.NewMsgUpdateClient(
		clientID, header,
		chain.SenderAccount.GetAddress(),
	)
	require.NoError(chain.t, err)

	return chain.sendMsgs(msg)
}

// ConstructUpdateTMClientHeader will construct a valid 07-tendermint Header to update the
// light client on the source chain.
func (chain *TestChain) ConstructUpdateTMClientHeader(counterparty *TestChain, clientID string) (*ibctmtypes.Header, error) {
	header := counterparty.LastHeader
	// Relayer must query for LatestHeight on client to get TrustedHeight
	trustedHeight := chain.GetClientState(clientID).GetLatestHeight().(clienttypes.Height)
	var (
		tmTrustedVals *tmtypes.ValidatorSet
		ok            bool
	)
	// Once we get TrustedHeight from client, we must query the validators from the counterparty chain
	// If the LatestHeight == LastHeader.Height, then TrustedValidators are current validators
	// If LatestHeight < LastHeader.Height, we can query the historical validator set from HistoricalInfo
	if trustedHeight == counterparty.LastHeader.GetHeight() {
		tmTrustedVals = counterparty.Vals
	} else {
		// NOTE: We need to get validators from counterparty at height: trustedHeight+1
		// since the last trusted validators for a header at height h
		// is the NextValidators at h+1 committed to in header h by
		// NextValidatorsHash
		tmTrustedVals, ok = counterparty.GetValsAtHeight(int64(trustedHeight.RevisionHeight + 1))
		if !ok {
			return nil, sdkerrors.Wrapf(ibctmtypes.ErrInvalidHeaderHeight, "could not retrieve trusted validators at trustedHeight: %d", trustedHeight)
		}
	}
	// inject trusted fields into last header
	// for now assume revision number is 0
	header.TrustedHeight = trustedHeight

	trustedVals, err := tmTrustedVals.ToProto()
	if err != nil {
		return nil, err
	}
	header.TrustedValidators = trustedVals

	return header, nil

}

// ExpireClient fast forwards the chain's block time by the provided amount of time which will
// expire any clients with a trusting period less than or equal to this amount of time.
func (chain *TestChain) ExpireClient(amount time.Duration) {
	chain.CurrentHeader.Time = chain.CurrentHeader.Time.Add(amount)
}

// CurrentTMClientHeader creates a TM header using the current header parameters
// on the chain. The trusted fields in the header are set to nil.
func (chain *TestChain) CurrentTMClientHeader() *ibctmtypes.Header {
	return chain.CreateTMClientHeader(chain.ChainID, chain.CurrentHeader.Height, clienttypes.Height{}, chain.CurrentHeader.Time, chain.Vals, nil, chain.Signers)
}

// CreateTMClientHeader creates a TM header to update the TM client. Args are passed in to allow
// caller flexibility to use params that differ from the chain.
func (chain *TestChain) CreateTMClientHeader(chainID string, blockHeight int64, trustedHeight clienttypes.Height, timestamp time.Time, tmValSet, tmTrustedVals *tmtypes.ValidatorSet, signers []tmtypes.PrivValidator) *ibctmtypes.Header {
	require.NotNil(chain.t, tmValSet)

	vsetHash := tmValSet.Hash()
	tmHeader := tmtypes.Header{
		Version:            tmprotoversion.Consensus{Block: tmversion.BlockProtocol, App: 2},
		ChainID:            chainID,
		Height:             blockHeight,
		Time:               timestamp,
		LastBlockID:        MakeBlockID(make([]byte, tmhash.Size), 10_000, make([]byte, tmhash.Size)),
		LastCommitHash:     chain.App.LastCommitID().Hash,
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: vsetHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            chain.CurrentHeader.AppHash,
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    tmValSet.Proposer.Address,
	}
	hhash := tmHeader.Hash()
	blockID := MakeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))
	voteSet := tmtypes.NewVoteSet(chainID, blockHeight, 1, tmproto.PrecommitType, tmValSet)

	commit, err := tmtypes.MakeCommit(blockID, blockHeight, 1, voteSet, signers, timestamp)
	require.NoError(chain.t, err)

	signedHeader := &tmproto.SignedHeader{
		Header: tmHeader.ToProto(),
		Commit: commit.ToProto(),
	}

	valSet, err := tmValSet.ToProto()
	if err != nil {
		panic(err)
	}
	var trustedVals *tmproto.ValidatorSet
	if tmTrustedVals != nil {
		trustedVals, err = tmTrustedVals.ToProto()
		if err != nil {
			panic(err)
		}
	}

	// The trusted fields may be nil. They may be filled before relaying messages to a client.
	// The relayer is responsible for querying client and injecting appropriate trusted fields.
	return &ibctmtypes.Header{
		SignedHeader:      signedHeader,
		ValidatorSet:      valSet,
		TrustedHeight:     trustedHeight,
		TrustedValidators: trustedVals,
	}
}

// MakeBlockID copied unimported test functions from tmtypes to use them here
func MakeBlockID(hash []byte, partSetSize uint32, partSetHash []byte) tmtypes.BlockID {
	return tmtypes.BlockID{
		Hash: hash,
		PartSetHeader: tmtypes.PartSetHeader{
			Total: partSetSize,
			Hash:  partSetHash,
		},
	}
}

// CreateSortedSignerArray takes two PrivValidators, and the corresponding Validator structs
// (including voting power). It returns a signer array of PrivValidators that matches the
// sorting of ValidatorSet.
// The sorting is first by .VotingPower (descending), with secondary index of .Address (ascending).
func CreateSortedSignerArray(altPrivVal, suitePrivVal tmtypes.PrivValidator,
	altVal, suiteVal *tmtypes.Validator) []tmtypes.PrivValidator {

	switch {
	case altVal.VotingPower > suiteVal.VotingPower:
		return []tmtypes.PrivValidator{altPrivVal, suitePrivVal}
	case altVal.VotingPower < suiteVal.VotingPower:
		return []tmtypes.PrivValidator{suitePrivVal, altPrivVal}
	default:
		if bytes.Compare(altVal.Address, suiteVal.Address) == -1 {
			return []tmtypes.PrivValidator{altPrivVal, suitePrivVal}
		}
		return []tmtypes.PrivValidator{suitePrivVal, altPrivVal}
	}
}

// ConnectionOpenInit will construct and execute a MsgConnectionOpenInit.
func (chain *TestChain) ConnectionOpenInit(
	counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
) error {
	msg := connectiontypes.NewMsgConnectionOpenInit(
		connection.ClientID,
		connection.CounterpartyClientID,
		counterparty.GetPrefix(), DefaultOpenInitVersion, DefaultDelayPeriod,
		chain.SenderAccount.GetAddress(),
	)
	return chain.sendMsgs(msg)
}

// ConnectionOpenTry will construct and execute a MsgConnectionOpenTry.
func (chain *TestChain) ConnectionOpenTry(
	counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
) error {
	counterpartyClient, proofClient := counterparty.QueryClientStateProof(counterpartyConnection.ClientID)

	connectionKey := host.ConnectionKey(counterpartyConnection.ID)
	proofInit, proofHeight := counterparty.QueryProof(connectionKey)

	proofConsensus, consensusHeight := counterparty.QueryConsensusStateProof(counterpartyConnection.ClientID)

	msg := connectiontypes.NewMsgConnectionOpenTry(
		"", connection.ClientID, // does not support handshake continuation
		counterpartyConnection.ID, counterpartyConnection.ClientID,
		counterpartyClient, counterparty.GetPrefix(), []*connectiontypes.Version{ConnectionVersion}, DefaultDelayPeriod,
		proofInit, proofClient, proofConsensus,
		proofHeight, consensusHeight,
		chain.SenderAccount.GetAddress(),
	)
	return chain.sendMsgs(msg)
}

// ConnectionOpenAck will construct and execute a MsgConnectionOpenAck.
func (chain *TestChain) ConnectionOpenAck(
	counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
) error {
	counterpartyClient, proofClient := counterparty.QueryClientStateProof(counterpartyConnection.ClientID)

	connectionKey := host.ConnectionKey(counterpartyConnection.ID)
	proofTry, proofHeight := counterparty.QueryProof(connectionKey)

	proofConsensus, consensusHeight := counterparty.QueryConsensusStateProof(counterpartyConnection.ClientID)

	msg := connectiontypes.NewMsgConnectionOpenAck(
		connection.ID, counterpartyConnection.ID, counterpartyClient, // testing doesn't use flexible selection
		proofTry, proofClient, proofConsensus,
		proofHeight, consensusHeight,
		ConnectionVersion,
		chain.SenderAccount.GetAddress(),
	)
	return chain.sendMsgs(msg)
}

// ConnectionOpenConfirm will construct and execute a MsgConnectionOpenConfirm.
func (chain *TestChain) ConnectionOpenConfirm(
	counterparty *TestChain,
	connection, counterpartyConnection *TestConnection,
) error {
	connectionKey := host.ConnectionKey(counterpartyConnection.ID)
	proof, height := counterparty.QueryProof(connectionKey)

	msg := connectiontypes.NewMsgConnectionOpenConfirm(
		connection.ID,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.sendMsgs(msg)
}

// CreatePortCapability binds and claims a capability for the given portID if it does not
// already exist. This function will fail testing on any resulting error.
// NOTE: only creation of a capbility for a transfer or mock port is supported
// Other applications must bind to the port in InitGenesis or modify this code.
func (chain *TestChain) CreatePortCapability(portID string) {
	// check if the portId is already binded, if not bind it
	_, ok := chain.App.ScopedIBCKeeper.GetCapability(chain.GetContext(), host.PortPath(portID))
	if !ok {
		// create capability using the IBC capability keeper
		cap, err := chain.App.ScopedIBCKeeper.NewCapability(chain.GetContext(), host.PortPath(portID))
		require.NoError(chain.t, err)

		switch portID {
		case MockPort:
			// claim capability using the mock capability keeper
			err = chain.App.ScopedIBCMockKeeper.ClaimCapability(chain.GetContext(), cap, host.PortPath(portID))
			require.NoError(chain.t, err)
		case TransferPort:
			// claim capability using the transfer capability keeper
			err = chain.App.ScopedTransferKeeper.ClaimCapability(chain.GetContext(), cap, host.PortPath(portID))
			require.NoError(chain.t, err)
		default:
			panic(fmt.Sprintf("unsupported ibc testing package port ID %s", portID))
		}
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
		ch.PortID,
		ch.Version, order, []string{connectionID},
		counterparty.PortID,
		chain.SenderAccount.GetAddress(),
	)
	return chain.sendMsgs(msg)
}

// ChanOpenTry will construct and execute a MsgChannelOpenTry.
func (chain *TestChain) ChanOpenTry(
	counterparty *TestChain,
	ch, counterpartyCh TestChannel,
	order channeltypes.Order,
	connectionID string,
) error {
	proof, height := counterparty.QueryProof(host.ChannelKey(counterpartyCh.PortID, counterpartyCh.ID))

	msg := channeltypes.NewMsgChannelOpenTry(
		ch.PortID, "", // does not support handshake continuation
		ch.Version, order, []string{connectionID},
		counterpartyCh.PortID, counterpartyCh.ID, counterpartyCh.Version,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.sendMsgs(msg)
}

// ChanOpenAck will construct and execute a MsgChannelOpenAck.
func (chain *TestChain) ChanOpenAck(
	counterparty *TestChain,
	ch, counterpartyCh TestChannel,
) error {
	proof, height := counterparty.QueryProof(host.ChannelKey(counterpartyCh.PortID, counterpartyCh.ID))

	msg := channeltypes.NewMsgChannelOpenAck(
		ch.PortID, ch.ID,
		counterpartyCh.ID, counterpartyCh.Version, // testing doesn't use flexible selection
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.sendMsgs(msg)
}

// ChanOpenConfirm will construct and execute a MsgChannelOpenConfirm.
func (chain *TestChain) ChanOpenConfirm(
	counterparty *TestChain,
	ch, counterpartyCh TestChannel,
) error {
	proof, height := counterparty.QueryProof(host.ChannelKey(counterpartyCh.PortID, counterpartyCh.ID))

	msg := channeltypes.NewMsgChannelOpenConfirm(
		ch.PortID, ch.ID,
		proof, height,
		chain.SenderAccount.GetAddress(),
	)
	return chain.sendMsgs(msg)
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
	return chain.sendMsgs(msg)
}

// GetPacketData returns a ibc-transfer marshalled packet to be used for
// callback testing.
func (chain *TestChain) GetPacketData(counterparty *TestChain) []byte {
	packet := ibctransfertypes.FungibleTokenPacketData{
		Denom:    TestCoin.Denom,
		Amount:   TestCoin.Amount.Uint64(),
		Sender:   chain.SenderAccount.GetAddress().String(),
		Receiver: counterparty.SenderAccount.GetAddress().String(),
	}

	return packet.GetBytes()
}

// SendPacket simulates sending a packet through the channel keeper. No message needs to be
// passed since this call is made from a module.
func (chain *TestChain) SendPacket(
	packet exported.PacketI,
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

// WriteAcknowledgement simulates writing an acknowledgement to the chain.
func (chain *TestChain) WriteAcknowledgement(
	packet exported.PacketI,
) error {
	channelCap := chain.GetChannelCapability(packet.GetDestPort(), packet.GetDestChannel())

	// no need to send message, acting as a handler
	err := chain.App.IBCKeeper.ChannelKeeper.WriteAcknowledgement(chain.GetContext(), channelCap, packet, TestHash)
	if err != nil {
		return err
	}

	// commit changes
	chain.App.Commit()
	chain.NextBlock()

	return nil
}
