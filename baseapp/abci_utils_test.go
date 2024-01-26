package baseapp_test

import (
	"bytes"
	"strings"
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
	ctrl := gomock.NewController(s.T())
	app := mock.NewMockProposalTxVerifier(ctrl)
	mp1 := mempool.NewPriorityMempool(
		mempool.PriorityNonceMempoolConfig[int64]{
			TxPriority:      mempool.NewDefaultTxPriority(),
			MaxTx:           0,
			SignerExtractor: mempool.NewDefaultSignerExtractionAdapter(),
		},
	)
	mp2 := mempool.NewPriorityMempool(
		mempool.PriorityNonceMempoolConfig[int64]{
			TxPriority:      mempool.NewDefaultTxPriority(),
			MaxTx:           0,
			SignerExtractor: mempool.NewDefaultSignerExtractionAdapter(),
		},
	)
	ph1 := baseapp.NewDefaultProposalHandler(mp1, app)
	handler1 := ph1.PrepareProposalHandler()
	ph2 := baseapp.NewDefaultProposalHandler(mp2, app)
	handler2 := ph2.PrepareProposalHandler()
	var (
		secret1 = []byte("secret1")
		secret2 = []byte("secret2")
		secret3 = []byte("secret3")
		secret4 = []byte("secret4")
		secret5 = []byte("secret5")
		secret6 = []byte("secret6")
		ctx1    = s.ctx.WithPriority(10)
		ctx2    = s.ctx.WithPriority(8)
	)

	tx1 := buildMsg(s.T(), txConfig, []byte(`1`), [][]byte{secret1}, []uint64{1})
	tx2 := buildMsg(s.T(), txConfig, []byte(`12345678910`), [][]byte{secret1}, []uint64{2})
	tx3 := buildMsg(s.T(), txConfig, []byte(`12`), [][]byte{secret1}, []uint64{3})
	tx4 := buildMsg(s.T(), txConfig, []byte(`12`), [][]byte{secret2}, []uint64{1})
	err := mp1.Insert(ctx1, tx1)
	s.Require().NoError(err)
	err = mp1.Insert(ctx1, tx2)
	s.Require().NoError(err)
	err = mp1.Insert(ctx1, tx3)
	s.Require().NoError(err)
	err = mp1.Insert(ctx2, tx4)
	s.Require().NoError(err)
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

	tx5 := buildMsg(s.T(), txConfig, []byte(`1`), [][]byte{secret1, secret2}, []uint64{1, 1})
	tx6 := buildMsg(s.T(), txConfig, []byte(`12345678910`), [][]byte{secret1, secret3}, []uint64{2, 1})
	tx7 := buildMsg(s.T(), txConfig, []byte(`12`), [][]byte{secret1, secret4}, []uint64{3, 1})
	tx8 := buildMsg(s.T(), txConfig, []byte(`12`), [][]byte{secret3, secret5}, []uint64{2, 1})
	tx9 := buildMsg(s.T(), txConfig, []byte(`12`), [][]byte{secret2, secret6}, []uint64{2, 1})

	err = mp2.Insert(ctx1, tx5)
	s.Require().NoError(err)
	err = mp2.Insert(ctx1, tx6)
	s.Require().NoError(err)
	err = mp2.Insert(ctx2, tx7)
	s.Require().NoError(err)
	err = mp2.Insert(ctx2, tx8)
	s.Require().NoError(err)
	err = mp2.Insert(ctx2, tx9)
	s.Require().NoError(err)
	txBz5, err := txConfig.TxEncoder()(tx5)
	s.Require().NoError(err)
	txBz6, err := txConfig.TxEncoder()(tx6)
	s.Require().NoError(err)
	txBz7, err := txConfig.TxEncoder()(tx7)
	s.Require().NoError(err)
	txBz8, err := txConfig.TxEncoder()(tx8)
	s.Require().NoError(err)
	txBz9, err := txConfig.TxEncoder()(tx9)
	s.Require().NoError(err)
	app.EXPECT().PrepareProposalVerifyTx(tx5).Return(txBz5, nil).AnyTimes()
	app.EXPECT().PrepareProposalVerifyTx(tx6).Return(txBz6, nil).AnyTimes()
	app.EXPECT().PrepareProposalVerifyTx(tx7).Return(txBz7, nil).AnyTimes()
	app.EXPECT().PrepareProposalVerifyTx(tx8).Return(txBz8, nil).AnyTimes()
	app.EXPECT().PrepareProposalVerifyTx(tx9).Return(txBz9, nil).AnyTimes()
	txDataSize5 := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz5}))
	txDataSize6 := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz6}))
	txDataSize7 := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz7}))
	txDataSize8 := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz8}))
	txDataSize9 := int(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz9}))
	s.Require().Equal(txDataSize5, 195)
	s.Require().Equal(txDataSize6, 205)
	s.Require().Equal(txDataSize7, 196)
	s.Require().Equal(txDataSize8, 196)
	s.Require().Equal(txDataSize9, 196)

	mapTxs := map[string]string{
		string(txBz1): "1",
		string(txBz2): "2",
		string(txBz3): "3",
		string(txBz4): "4",
		string(txBz5): "5",
		string(txBz6): "6",
		string(txBz7): "7",
		string(txBz8): "8",
		string(txBz9): "9",
	}
	testCases := map[string]struct {
		ctx         sdk.Context
		req         *abci.RequestPrepareProposal
		handler     sdk.PrepareProposalHandler
		expectedTxs [][]byte
	}{
		"skip same-sender non-sequential sequence and then add others txs": {
			ctx: s.ctx,
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz1, txBz2, txBz3, txBz4},
				MaxTxBytes: 111 + 112,
			},
			handler:     handler1,
			expectedTxs: [][]byte{txBz1, txBz4},
		},
		"skip multi-signers msg non-sequential sequence": {
			ctx: s.ctx,
			req: &abci.RequestPrepareProposal{
				Txs:        [][]byte{txBz5, txBz6, txBz7, txBz8, txBz9},
				MaxTxBytes: 195 + 196,
			},
			handler:     handler2,
			expectedTxs: [][]byte{txBz5, txBz9},
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			resp, err := tc.handler(tc.ctx, tc.req)
			s.Require().NoError(err)
			s.Require().EqualValues(toHumanReadable(mapTxs, resp.Txs), toHumanReadable(mapTxs, tc.expectedTxs))
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

func toHumanReadable(mapTxs map[string]string, txs [][]byte) string {
	strs := []string{}
	for _, v := range txs {
		strs = append(strs, mapTxs[string(v)])
	}
	return strings.Join(strs, ",")
}

func setTxSignatureWithSecret(t *testing.T, builder client.TxBuilder, signatures ...signingtypes.SignatureV2) {
	t.Helper()
	err := builder.SetSignatures(
		signatures...,
	)
	require.NoError(t, err)
}
