package baseapp_test

import (
	"bytes"
	"sort"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtsecp256k1 "github.com/cometbft/cometbft/crypto/secp256k1"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/baseapp/testutil/mock"
	"github.com/cosmos/cosmos-sdk/client"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

const (
	chainID = "chain-id"
)

type testValidator struct {
	consAddr sdk.ConsAddress
	tmPk     cmtprotocrypto.PublicKey
	privKey  cmtsecp256k1.PrivKey
}

func newTestValidator() testValidator {
	privkey := cmtsecp256k1.GenPrivKey()
	pubkey := privkey.PubKey()
	tmPk := cmtprotocrypto.PublicKey{
		Sum: &cmtprotocrypto.PublicKey_Secp256K1{
			Secp256K1: pubkey.Bytes(),
		},
	}

	return testValidator{
		consAddr: sdk.ConsAddress(pubkey.Address()),
		tmPk:     tmPk,
		privKey:  privkey,
	}
}

func (t testValidator) toValidator(power int64) abci.Validator {
	return abci.Validator{
		Address: t.consAddr.Bytes(),
		Power:   power,
	}
}

type ABCIUtilsTestSuite struct {
	suite.Suite

	valStore *mock.MockValidatorStore
	vals     [3]testValidator
	ctx      sdk.Context
}

func NewABCIUtilsTestSuite(t *testing.T) *ABCIUtilsTestSuite {
	t.Helper()
	// create 3 validators
	s := &ABCIUtilsTestSuite{
		vals: [3]testValidator{
			newTestValidator(),
			newTestValidator(),
			newTestValidator(),
		},
	}

	// create mock
	ctrl := gomock.NewController(t)
	valStore := mock.NewMockValidatorStore(ctrl)
	s.valStore = valStore

	// set up mock
	s.valStore.EXPECT().GetPubKeyByConsAddr(gomock.Any(), s.vals[0].consAddr.Bytes()).Return(s.vals[0].tmPk, nil).AnyTimes()
	s.valStore.EXPECT().GetPubKeyByConsAddr(gomock.Any(), s.vals[1].consAddr.Bytes()).Return(s.vals[1].tmPk, nil).AnyTimes()
	s.valStore.EXPECT().GetPubKeyByConsAddr(gomock.Any(), s.vals[2].consAddr.Bytes()).Return(s.vals[2].tmPk, nil).AnyTimes()

	// create context
	s.ctx = sdk.Context{}.WithConsensusParams(cmtproto.ConsensusParams{
		Abci: &cmtproto.ABCIParams{
			VoteExtensionsEnableHeight: 2,
		},
	}).WithBlockHeader(cmtproto.Header{
		ChainID: chainID,
	}).WithLogger(log.NewTestLogger(t))
	return s
}

func TestABCIUtilsTestSuite(t *testing.T) {
	suite.Run(t, NewABCIUtilsTestSuite(t))
}

// check ValidateVoteExtensions works when all nodes have CommitBlockID votes
func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsHappyPath() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	extSig1, err := s.vals[1].privKey.Sign(bz)
	s.Require().NoError(err)

	extSig2, err := s.vals[2].privKey.Sign(bz)
	s.Require().NoError(err)

	s.ctx = s.ctx.WithBlockHeight(3).WithHeaderInfo(header.Info{Height: 3, ChainID: chainID}) // enable vote-extensions

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			{
				Validator:          s.vals[0].toValidator(333),
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:          s.vals[1].toValidator(333),
				VoteExtension:      ext,
				ExtensionSignature: extSig1,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:          s.vals[2].toValidator(334),
				VoteExtension:      ext,
				ExtensionSignature: extSig2,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
		},
	}

	// order + convert to last commit
	llc, info := extendedCommitToLastCommit(llc)
	s.ctx = s.ctx.WithCometInfo(info)

	// expect-pass (votes of height 2 are included in next block)
	s.Require().NoError(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 0, "", llc))
}

// check ValidateVoteExtensions works when a single node has submitted a BlockID_Absent
func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsSingleVoteAbsent() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	extSig2, err := s.vals[2].privKey.Sign(bz)
	s.Require().NoError(err)

	s.ctx = s.ctx.WithBlockHeight(3).WithHeaderInfo(header.Info{Height: 3, ChainID: chainID}) // vote-extensions are enabled

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			{
				Validator:          s.vals[0].toValidator(333),
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
			// validator of power <1/3 is missing, so commit-info shld still be valid
			{
				Validator:   s.vals[1].toValidator(333),
				BlockIdFlag: cmtproto.BlockIDFlagAbsent,
			},
			{
				Validator:          s.vals[2].toValidator(334),
				VoteExtension:      ext,
				ExtensionSignature: extSig2,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
		},
	}

	llc, info := extendedCommitToLastCommit(llc)
	s.ctx = s.ctx.WithCometInfo(info)

	// expect-pass (votes of height 2 are included in next block)
	s.Require().NoError(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 0, "", llc))
}

// check ValidateVoteExtensions works with duplicate votes
func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsDuplicateVotes() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	ve := abci.ExtendedVoteInfo{
		Validator:          s.vals[0].toValidator(333),
		VoteExtension:      ext,
		ExtensionSignature: extSig0,
		BlockIdFlag:        cmtproto.BlockIDFlagCommit,
	}

	ve2 := abci.ExtendedVoteInfo{
		Validator:          s.vals[0].toValidator(334), // use diff voting-power to dupe
		VoteExtension:      ext,
		ExtensionSignature: extSig0,
		BlockIdFlag:        cmtproto.BlockIDFlagCommit,
	}

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			ve,
			ve2,
		},
	}

	s.ctx = s.ctx.WithBlockHeight(3).WithHeaderInfo(header.Info{Height: 3, ChainID: chainID}) // vote-extensions are enabled
	llc, info := extendedCommitToLastCommit(llc)
	s.ctx = s.ctx.WithCometInfo(info)

	// expect fail (duplicate votes)
	s.Require().Error(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 0, "", llc))
}

// check ValidateVoteExtensions works when a single node has submitted a BlockID_Nil
func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsSingleVoteNil() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	extSig2, err := s.vals[2].privKey.Sign(bz)
	s.Require().NoError(err)

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			{
				Validator:          s.vals[0].toValidator(333),
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
			// validator of power <1/3 is missing, so commit-info should still be valid
			{
				Validator:   s.vals[1].toValidator(333),
				BlockIdFlag: cmtproto.BlockIDFlagNil,
			},
			{
				Validator:          s.vals[2].toValidator(334),
				VoteExtension:      ext,
				ExtensionSignature: extSig2,
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
			},
		},
	}

	s.ctx = s.ctx.WithBlockHeight(3).WithHeaderInfo(header.Info{Height: 3, ChainID: chainID}) // vote-extensions are enabled

	// create last commit
	llc, info := extendedCommitToLastCommit(llc)
	s.ctx = s.ctx.WithCometInfo(info)

	// expect-pass (votes of height 2 are included in next block)
	s.Require().NoError(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 0, "", llc))
}

// check ValidateVoteExtensions works when two nodes have submitted a BlockID_Nil / BlockID_Absent
func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsTwoVotesNilAbsent() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			// validator of power >2/3 is missing, so commit-info should not be valid
			{
				Validator:          s.vals[0].toValidator(333),
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
			},
			{
				Validator:   s.vals[1].toValidator(333),
				BlockIdFlag: cmtproto.BlockIDFlagNil,
			},
			{
				Validator:     s.vals[2].toValidator(334),
				VoteExtension: ext,
				BlockIdFlag:   cmtproto.BlockIDFlagAbsent,
			},
		},
	}

	s.ctx = s.ctx.WithBlockHeight(3).WithHeaderInfo(header.Info{Height: 3, ChainID: chainID}) // vote-extensions are enabled

	// create last commit
	llc, info := extendedCommitToLastCommit(llc)
	s.ctx = s.ctx.WithCometInfo(info)

	// expect-pass (votes of height 2 are included in next block)
	s.Require().Error(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 0, "", llc))
}

func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsIncorrectVotingPower() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			// validator of power >2/3 is missing, so commit-info should not be valid
			{
				Validator:          s.vals[0].toValidator(333),
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
			},
			{
				Validator:   s.vals[1].toValidator(333),
				BlockIdFlag: cmtproto.BlockIDFlagNil,
			},
			{
				Validator:     s.vals[2].toValidator(334),
				VoteExtension: ext,
				BlockIdFlag:   cmtproto.BlockIDFlagAbsent,
			},
		},
	}

	s.ctx = s.ctx.WithBlockHeight(3).WithHeaderInfo(header.Info{Height: 3, ChainID: chainID}) // vote-extensions are enabled

	// create last commit
	llc, info := extendedCommitToLastCommit(llc)
	s.ctx = s.ctx.WithCometInfo(info)

	// modify voting powers to differ from the last-commit
	llc.Votes[0].Validator.Power = 335
	llc.Votes[2].Validator.Power = 332

	// expect-pass (votes of height 2 are included in next block)
	s.Require().Error(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 0, "", llc))
}

func (s *ABCIUtilsTestSuite) TestValidateVoteExtensionsIncorrectOrder() {
	ext := []byte("vote-extension")
	cve := cmtproto.CanonicalVoteExtension{
		Extension: ext,
		Height:    2,
		Round:     int64(0),
		ChainId:   chainID,
	}

	bz, err := marshalDelimitedFn(&cve)
	s.Require().NoError(err)

	extSig0, err := s.vals[0].privKey.Sign(bz)
	s.Require().NoError(err)

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			// validator of power >2/3 is missing, so commit-info should not be valid
			{
				Validator:          s.vals[0].toValidator(333),
				BlockIdFlag:        cmtproto.BlockIDFlagCommit,
				VoteExtension:      ext,
				ExtensionSignature: extSig0,
			},
			{
				Validator:   s.vals[1].toValidator(333),
				BlockIdFlag: cmtproto.BlockIDFlagNil,
			},
			{
				Validator:     s.vals[2].toValidator(334),
				VoteExtension: ext,
				BlockIdFlag:   cmtproto.BlockIDFlagAbsent,
			},
		},
	}

	s.ctx = s.ctx.WithBlockHeight(3).WithHeaderInfo(header.Info{Height: 3, ChainID: chainID}) // vote-extensions are enabled

	// create last commit
	llc, info := extendedCommitToLastCommit(llc)
	s.ctx = s.ctx.WithCometInfo(info)

	// modify voting powers to differ from the last-commit
	llc.Votes[0], llc.Votes[2] = llc.Votes[2], llc.Votes[0]

	// expect-pass (votes of height 2 are included in next block)
	s.Require().Error(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 0, "", llc))
}

func (s *ABCIUtilsTestSuite) TestDefaultProposalHandler_NoOpMempoolTxSelection() {
	// create a codec for marshaling
	cdc := codectestutil.CodecOptions{}.NewCodec()
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())

	// create a baseapp along with a tx config for tx generation
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	app := baseapp.NewBaseApp(s.T().Name(), log.NewNopLogger(), dbm.NewMemDB(), txConfig.TxDecoder())

	// create a proposal handler
	ph := baseapp.NewDefaultProposalHandler(mempool.NoOpMempool{}, app)
	handler := ph.PrepareProposalHandler()

	// build a tx
	_, _, addr := testdata.KeyTestPubAddr()
	builder := txConfig.NewTxBuilder()
	s.Require().NoError(builder.SetMsgs(
		&baseapptestutil.MsgCounter{Counter: 0, FailOnHandler: false, Signer: addr.String()},
	))
	builder.SetGasLimit(100)
	setTxSignature(s.T(), builder, 0)

	// encode the tx to be used in the proposal request
	tx := builder.GetTx()
	txBz, err := txConfig.TxEncoder()(tx)
	s.Require().NoError(err)
	s.Require().Len(txBz, 152)

	testCases := map[string]struct {
		ctx         sdk.Context
		req         *abci.RequestPrepareProposal
		expectedTxs int
	}{
		"small max tx bytes": {
			ctx: s.ctx,
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz, txBz, txBz, txBz, txBz},
				MaxTxBytes: 10,
			},
			expectedTxs: 0,
		},
		"small max gas": {
			ctx: s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
				Block: &cmtproto.BlockParams{
					MaxGas: 10,
				},
			}),
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz, txBz, txBz, txBz, txBz},
				MaxTxBytes: 456,
			},
			expectedTxs: 0,
		},
		"large max tx bytes": {
			ctx: s.ctx,
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz, txBz, txBz, txBz, txBz},
				MaxTxBytes: 456,
			},
			expectedTxs: 3,
		},
		"max gas and tx bytes": {
			ctx: s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
				Block: &cmtproto.BlockParams{
					MaxGas: 200,
				},
			}),
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz, txBz, txBz, txBz, txBz},
				MaxTxBytes: 456,
			},
			expectedTxs: 2,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			// iterate multiple times to ensure the tx selector is cleared each time
			for i := 0; i < 5; i++ {
				resp, err := handler(tc.ctx, tc.req)
				s.Require().NoError(err)
				s.Require().Len(resp.Txs, tc.expectedTxs)
			}
		})
	}
}

func (s *ABCIUtilsTestSuite) TestDefaultProposalHandler_PriorityNonceMempoolTxSelection() {
	cdc := codectestutil.CodecOptions{}.NewCodec()
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	var (
		secret1 = []byte("secret1")
		secret2 = []byte("secret2")
		secret3 = []byte("secret3")
		secret4 = []byte("secret4")
		secret5 = []byte("secret5")
		secret6 = []byte("secret6")
	)

	type testTx struct {
		tx       sdk.Tx
		priority int64
		bz       []byte
		size     int
	}

	testTxs := []testTx{
		// test 1
		{tx: buildMsg(s.T(), txConfig, []byte(`0`), [][]byte{secret1}, []uint64{1}), priority: 10},
		{tx: buildMsg(s.T(), txConfig, []byte(`12345678910`), [][]byte{secret1}, []uint64{2}), priority: 10},
		{tx: buildMsg(s.T(), txConfig, []byte(`22`), [][]byte{secret1}, []uint64{3}), priority: 10},
		{tx: buildMsg(s.T(), txConfig, []byte(`32`), [][]byte{secret2}, []uint64{1}), priority: 8},
		// test 2
		{tx: buildMsg(s.T(), txConfig, []byte(`4`), [][]byte{secret1, secret2}, []uint64{3, 3}), priority: 10},
		{tx: buildMsg(s.T(), txConfig, []byte(`52345678910`), [][]byte{secret1, secret3}, []uint64{4, 3}), priority: 10},
		{tx: buildMsg(s.T(), txConfig, []byte(`62`), [][]byte{secret1, secret4}, []uint64{5, 3}), priority: 8},
		{tx: buildMsg(s.T(), txConfig, []byte(`72`), [][]byte{secret3, secret5}, []uint64{4, 3}), priority: 8},
		{tx: buildMsg(s.T(), txConfig, []byte(`82`), [][]byte{secret2, secret6}, []uint64{4, 3}), priority: 8},
		// test 3
		{tx: buildMsg(s.T(), txConfig, []byte(`9`), [][]byte{secret3, secret4}, []uint64{3, 3}), priority: 10},
		{tx: buildMsg(s.T(), txConfig, []byte(`1052345678910`), [][]byte{secret1, secret2}, []uint64{4, 4}), priority: 8},
		{tx: buildMsg(s.T(), txConfig, []byte(`11`), [][]byte{secret1, secret2}, []uint64{5, 5}), priority: 8},
		// test 4
		{tx: buildMsg(s.T(), txConfig, []byte(`1252345678910`), [][]byte{secret1}, []uint64{3}), priority: 10},
		{tx: buildMsg(s.T(), txConfig, []byte(`13`), [][]byte{secret1}, []uint64{5}), priority: 10},
		{tx: buildMsg(s.T(), txConfig, []byte(`14`), [][]byte{secret1}, []uint64{6}), priority: 8},
	}

	for i := range testTxs {
		bz, err := txConfig.TxEncoder()(testTxs[i].tx)
		s.Require().NoError(err)
		testTxs[i].bz = bz
		testTxs[i].size = int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{bz}))
	}

	s.Require().Equal(testTxs[0].size, 111)
	s.Require().Equal(testTxs[1].size, 121)
	s.Require().Equal(testTxs[2].size, 112)
	s.Require().Equal(testTxs[3].size, 112)
	s.Require().Equal(testTxs[4].size, 195)
	s.Require().Equal(testTxs[5].size, 205)
	s.Require().Equal(testTxs[6].size, 196)
	s.Require().Equal(testTxs[7].size, 196)
	s.Require().Equal(testTxs[8].size, 196)

	testCases := map[string]struct {
		ctx         sdk.Context
		txInputs    []testTx
		req         *abci.RequestPrepareProposal
		handler     sdk.PrepareProposalHandler
		expectedTxs []int
	}{
		"skip same-sender non-sequential sequence and then add others txs": {
			ctx:      s.ctx,
			txInputs: []testTx{testTxs[0], testTxs[1], testTxs[2], testTxs[3]},
			req: &abci.RequestPrepareProposal{
				MaxTxBytes: 111 + 112,
			},
			expectedTxs: []int{0, 3},
		},
		"skip multi-signers msg non-sequential sequence": {
			ctx:      s.ctx,
			txInputs: []testTx{testTxs[4], testTxs[5], testTxs[6], testTxs[7], testTxs[8]},
			req: &abci.RequestPrepareProposal{
				MaxTxBytes: 195 + 196,
			},
			expectedTxs: []int{4, 8},
		},
		"only the first tx is added": {
			// Because tx 10 is valid, tx 11 can't be valid as they have higher sequence numbers.
			ctx:      s.ctx,
			txInputs: []testTx{testTxs[9], testTxs[10], testTxs[11]},
			req: &abci.RequestPrepareProposal{
				MaxTxBytes: 195 + 196,
			},
			expectedTxs: []int{9},
		},
		"no txs added": {
			// Becasuse the first tx was deemed valid but too big, the next expected valid sequence is tx[0].seq (3), so
			// the rest of the txs fail because they have a seq of 4.
			ctx:      s.ctx,
			txInputs: []testTx{testTxs[12], testTxs[13], testTxs[14]},
			req: &abci.RequestPrepareProposal{
				MaxTxBytes: 112,
			},
			expectedTxs: []int{},
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			ctrl := gomock.NewController(s.T())
			app := mock.NewMockProposalTxVerifier(ctrl)
			mp := mempool.NewPriorityMempool(
				mempool.PriorityNonceMempoolConfig[int64]{
					TxPriority:      mempool.NewDefaultTxPriority(),
					MaxTx:           0,
					SignerExtractor: mempool.NewDefaultSignerExtractionAdapter(),
				},
			)

			ph := baseapp.NewDefaultProposalHandler(mp, app)

			for _, v := range tc.txInputs {
				app.EXPECT().PrepareProposalVerifyTx(v.tx).Return(v.bz, nil).AnyTimes()
				s.NoError(mp.Insert(s.ctx.WithPriority(v.priority), v.tx))
				tc.req.Txs = append(tc.req.Txs, v.bz)
			}

			resp, err := ph.PrepareProposalHandler()(tc.ctx, tc.req)
			s.Require().NoError(err)
			respTxIndexes := []int{}
			for _, tx := range resp.Txs {
				for i, v := range testTxs {
					if bytes.Equal(tx, v.bz) {
						respTxIndexes = append(respTxIndexes, i)
					}
				}
			}

			s.Require().EqualValues(tc.expectedTxs, respTxIndexes)
		})
	}
}

func marshalDelimitedFn(msg proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func buildMsg(t *testing.T, txConfig client.TxConfig, value []byte, secrets [][]byte, nonces []uint64) sdk.Tx {
	t.Helper()
	builder := txConfig.NewTxBuilder()
	_ = builder.SetMsgs(
		&baseapptestutil.MsgKeyValue{Value: value},
	)
	require.Equal(t, len(secrets), len(nonces))
	signatures := make([]signingtypes.SignatureV2, 0)
	for index, secret := range secrets {
		nonce := nonces[index]
		privKey := secp256k1.GenPrivKeyFromSecret(secret)
		pubKey := privKey.PubKey()
		signatures = append(signatures, signingtypes.SignatureV2{
			PubKey:   pubKey,
			Sequence: nonce,
			Data:     &signingtypes.SingleSignatureData{},
		})
	}
	setTxSignatureWithSecret(t, builder, signatures...)
	return builder.GetTx()
}

func setTxSignatureWithSecret(t *testing.T, builder client.TxBuilder, signatures ...signingtypes.SignatureV2) {
	t.Helper()
	err := builder.SetSignatures(
		signatures...,
	)
	require.NoError(t, err)
}

func extendedCommitToLastCommit(ec abci.ExtendedCommitInfo) (abci.ExtendedCommitInfo, comet.BlockInfo) {
	// sort the extended commit info
	sort.Sort(extendedVoteInfos(ec.Votes))

	// convert the extended commit info to last commit info
	lastCommit := abci.CommitInfo{
		Round: ec.Round,
		Votes: make([]abci.VoteInfo, len(ec.Votes)),
	}

	for i, vote := range ec.Votes {
		lastCommit.Votes[i] = abci.VoteInfo{
			Validator: abci.Validator{
				Address: vote.Validator.Address,
				Power:   vote.Validator.Power,
			},
		}
	}

	return ec, baseapp.NewBlockInfo(
		nil,
		nil,
		nil,
		lastCommit,
	)
}

type extendedVoteInfos []abci.ExtendedVoteInfo

func (v extendedVoteInfos) Len() int {
	return len(v)
}

func (v extendedVoteInfos) Less(i, j int) bool {
	if v[i].Validator.Power == v[j].Validator.Power {
		return bytes.Compare(v[i].Validator.Address, v[j].Validator.Address) == -1
	}
	return v[i].Validator.Power > v[j].Validator.Power
}

func (v extendedVoteInfos) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
