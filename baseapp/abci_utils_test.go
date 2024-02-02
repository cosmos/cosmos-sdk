package baseapp_test

import (
	"bytes"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtsecp256k1 "github.com/cometbft/cometbft/crypto/secp256k1"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/baseapp/testutil/mock"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
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

	vals [3]testValidator
	ctx  sdk.Context
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

	// create context
	s.ctx = sdk.Context{}.WithConsensusParams(&cmtproto.ConsensusParams{})
	return s
}

func TestABCIUtilsTestSuite(t *testing.T) {
	suite.Run(t, NewABCIUtilsTestSuite(t))
}

func (s *ABCIUtilsTestSuite) TestDefaultProposalHandler_PriorityNonceMempoolTxSelection() {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
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
		req         abci.RequestPrepareProposal
		handler     sdk.PrepareProposalHandler
		expectedTxs []int
	}{
		"skip same-sender non-sequential sequence and then add others txs": {
			ctx:      s.ctx,
			txInputs: []testTx{testTxs[0], testTxs[1], testTxs[2], testTxs[3]},
			req: abci.RequestPrepareProposal{
				MaxTxBytes: 111 + 112,
			},
			expectedTxs: []int{0, 3},
		},
		"skip multi-signers msg non-sequential sequence": {
			ctx:      s.ctx,
			txInputs: []testTx{testTxs[4], testTxs[5], testTxs[6], testTxs[7], testTxs[8]},
			req: abci.RequestPrepareProposal{
				MaxTxBytes: 195 + 196,
			},
			expectedTxs: []int{4, 8},
		},
		"only the first tx is added": {
			// Because tx 10 is valid, tx 11 can't be valid as they have higher sequence numbers.
			ctx:      s.ctx,
			txInputs: []testTx{testTxs[9], testTxs[10], testTxs[11]},
			req: abci.RequestPrepareProposal{
				MaxTxBytes: 195 + 196,
			},
			expectedTxs: []int{9},
		},
		"no txs added": {
			// Becasuse the first tx was deemed valid but too big, the next expected valid sequence is tx[0].seq (3), so
			// the rest of the txs fail because they have a seq of 4.
			ctx:      s.ctx,
			txInputs: []testTx{testTxs[12], testTxs[13], testTxs[14]},
			req: abci.RequestPrepareProposal{
				MaxTxBytes: 112,
			},
			expectedTxs: []int{},
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			ctrl := gomock.NewController(s.T())
			app := mock.NewMockProposalTxVerifier(ctrl)
			mp := mempool.NewPriorityMempool()
			ph := baseapp.NewDefaultProposalHandler(mp, app).PrepareProposalHandler()

			for _, v := range tc.txInputs {
				app.EXPECT().PrepareProposalVerifyTx(v.tx).Return(v.bz, nil).AnyTimes()
				s.NoError(mp.Insert(s.ctx.WithPriority(v.priority), v.tx))
				tc.req.Txs = append(tc.req.Txs, v.bz)
			}

			resp := ph(tc.ctx, tc.req)
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
