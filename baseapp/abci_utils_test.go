package baseapp_test

import (
	"bytes"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	authtx "cosmossdk.io/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/baseapp/testutil/mock"
	"github.com/cosmos/cosmos-sdk/client"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)

const (
	chainID = "chain-id"
)

type testValidator struct {
	consAddr sdk.ConsAddress
	tmPk     cmtprotocrypto.PublicKey
	privKey  secp256k1.PrivKey
}

func newTestValidator() testValidator {
	privkey := secp256k1.GenPrivKey()
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
	})
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
	// expect-pass (votes of height 2 are included in next block)
	s.Require().NoError(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 3, chainID, llc))
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
	// expect-pass (votes of height 2 are included in next block)
	s.Require().NoError(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 3, chainID, llc))
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

	llc := abci.ExtendedCommitInfo{
		Round: 0,
		Votes: []abci.ExtendedVoteInfo{
			ve,
			ve,
		},
	}
	// expect fail (duplicate votes)
	s.Require().Error(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 3, chainID, llc))
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
	// expect-pass (votes of height 2 are included in next block)
	s.Require().NoError(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 3, chainID, llc))
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

	// expect-pass (votes of height 2 are included in next block)
	s.Require().Error(baseapp.ValidateVoteExtensions(s.ctx, s.valStore, 3, chainID, llc))
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

	txDataSize := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz}))
	s.Require().Equal(txDataSize, 155)

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
				MaxTxBytes: 465,
			},
			expectedTxs: 0,
		},
		"large max tx bytes": {
			ctx: s.ctx,
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz, txBz, txBz, txBz, txBz},
				MaxTxBytes: 465,
			},
			expectedTxs: 3,
		},
		"large max tx bytes len calculation": {
			ctx: s.ctx,
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz, txBz, txBz, txBz, txBz},
				MaxTxBytes: 456,
			},
			expectedTxs: 2,
		},
		"max gas and tx bytes": {
			ctx: s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
				Block: &cmtproto.BlockParams{
					MaxGas: 200,
				},
			}),
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz, txBz, txBz, txBz, txBz},
				MaxTxBytes: 465,
			},
			expectedTxs: 2,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			// iterate multiple times to ensure the tx selector is cleared each time
			for i := 0; i < 6; i++ {
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
	handler := ph.PrepareProposalHandler()
	var (
		secret1 = []byte("secret1")
		secret2 = []byte("secret2")
		tx1     = buildMsg(s.T(), txConfig, []byte(`1`), secret1, 1)
		ctx1    = s.ctx.WithPriority(10)
		tx2     = buildMsg(s.T(), txConfig, []byte(`12345678910`), secret1, 2)
		tx3     = buildMsg(s.T(), txConfig, []byte(`12`), secret1, 3)

		ctx2 = s.ctx.WithPriority(8)
		tx4  = buildMsg(s.T(), txConfig, []byte(`12`), secret2, 1)
	)
	mp.Insert(ctx1, tx1)
	mp.Insert(ctx1, tx2)
	mp.Insert(ctx1, tx3)
	mp.Insert(ctx2, tx4)

	txBz1, err := txConfig.TxEncoder()(tx1)
	s.Require().NoError(err)
	txBz2, err := txConfig.TxEncoder()(tx2)
	s.Require().NoError(err)
	txBz3, err := txConfig.TxEncoder()(tx3)
	s.Require().NoError(err)
	txBz4, err := txConfig.TxEncoder()(tx4)
	s.Require().NoError(err)

	app.EXPECT().PrepareProposalVerifyTx(tx1).Return(txBz1, nil).AnyTimes()
	app.EXPECT().PrepareProposalVerifyTx(tx2).Return(txBz2, nil).AnyTimes()
	app.EXPECT().PrepareProposalVerifyTx(tx3).Return(txBz3, nil).AnyTimes()
	app.EXPECT().PrepareProposalVerifyTx(tx4).Return(txBz4, nil).AnyTimes()

	txDataSize1 := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz1}))
	txDataSize2 := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz2}))
	txDataSize3 := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz3}))
	txDataSize4 := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz4}))
	s.Require().Equal(txDataSize1, 111)
	s.Require().Equal(txDataSize2, 121)
	s.Require().Equal(txDataSize3, 112)
	s.Require().Equal(txDataSize4, 112)

	testCases := map[string]struct {
		ctx         sdk.Context
		req         *abci.RequestPrepareProposal
		expectedTxs [][]byte
	}{
		"must skip same-sender non-sequential sequence": {
			ctx: s.ctx,
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz1, txBz2, txBz3, txBz4},
				MaxTxBytes: 111 + 112 + 1,
			},
			expectedTxs: [][]byte{txBz1},
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			resp, err := handler(tc.ctx, tc.req)
			s.Require().NoError(err)
			s.Require().EqualValues(resp.Txs, tc.expectedTxs)
		})
	}
}

func buildMsg(t *testing.T, txConfig client.TxConfig, value, secret []byte, nonce uint64) sdk.Tx {
	builder := txConfig.NewTxBuilder()
	builder.SetMsgs(
		&baseapptestutil.MsgKeyValue{Value: value},
	)
	setTxSignatureWithSecret(t, builder, nonce, secret)
	return builder.GetTx()
}

func marshalDelimitedFn(msg proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
