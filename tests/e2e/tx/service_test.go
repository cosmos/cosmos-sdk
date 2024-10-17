package tx_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network network.NetworkI

	txHeight    int64
	queryClient tx.ServiceClient
	goodTxHash  string
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	s.cfg = cfg

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	val := s.network.GetValidators()[0]
	s.Require().NoError(s.network.WaitForNextBlock())

	s.queryClient = tx.NewServiceClient(val.GetClientCtx())

	msgSend := &banktypes.MsgSend{
		FromAddress: val.GetAddress().String(),
		ToAddress:   val.GetAddress().String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))),
	}

	// Create a new MsgSend tx from val to itself.
	out, err := cli.SubmitTestTx(
		val.GetClientCtx(),
		msgSend,
		val.GetAddress(),
		cli.TestTxConfig{
			Memo: "foobar",
		},
	)

	s.Require().NoError(err)

	var txRes sdk.TxResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().Equal(uint32(0), txRes.Code, txRes)
	s.goodTxHash = txRes.TxHash

	msgSend1 := &banktypes.MsgSend{
		FromAddress: val.GetAddress().String(),
		ToAddress:   val.GetAddress().String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1))),
	}

	out1, err := cli.SubmitTestTx(
		val.GetClientCtx(),
		msgSend1,
		val.GetAddress(),
		cli.TestTxConfig{
			Offline: true,
			AccNum:  0,
			Seq:     2,
			Memo:    "foobar",
		},
	)

	s.Require().NoError(err)
	var tr sdk.TxResponse
	s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(out1.Bytes(), &tr))
	s.Require().Equal(uint32(0), tr.Code)

	resp, err := cli.GetTxResponse(s.network, val.GetClientCtx(), tr.TxHash)
	s.Require().NoError(err)
	s.txHeight = resp.Height
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) readTestAminoTxJSON() ([]byte, *legacytx.StdTx) {
	val := s.network.GetValidators()[0]
	txJSONBytes, err := os.ReadFile("testdata/tx_amino1.json")
	s.Require().NoError(err)
	var stdTx legacytx.StdTx
	err = val.GetClientCtx().LegacyAmino.UnmarshalJSON(txJSONBytes, &stdTx)
	s.Require().NoError(err)
	return txJSONBytes, &stdTx
}

func (s *E2ETestSuite) TestTxEncodeAmino_GRPC() {
	val := s.network.GetValidators()[0]
	txJSONBytes, stdTx := s.readTestAminoTxJSON()

	testCases := []struct {
		name      string
		req       *tx.TxEncodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxEncodeAminoRequest{}, true, "invalid empty tx json"},
		{"invalid request", &tx.TxEncodeAminoRequest{AminoJson: "invalid tx json"}, true, "invalid request"},
		{"valid request with amino-json", &tx.TxEncodeAminoRequest{AminoJson: string(txJSONBytes)}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.queryClient.TxEncodeAmino(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
				s.Require().Empty(res)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(res.GetAminoBinary())

				var decodedTx legacytx.StdTx
				err = val.GetClientCtx().LegacyAmino.Unmarshal(res.AminoBinary, &decodedTx)
				s.Require().NoError(err)
				s.Require().Equal(decodedTx.GetMsgs(), stdTx.GetMsgs())
			}
		})
	}
}

func (s *E2ETestSuite) TestTxEncodeAmino_GRPCGateway() {
	val := s.network.GetValidators()[0]
	txJSONBytes, stdTx := s.readTestAminoTxJSON()

	testCases := []struct {
		name      string
		req       *tx.TxEncodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxEncodeAminoRequest{}, true, "invalid empty tx json"},
		{"invalid request", &tx.TxEncodeAminoRequest{AminoJson: "invalid tx json"}, true, "invalid request"},
		{"valid request with amino-json", &tx.TxEncodeAminoRequest{AminoJson: string(txJSONBytes)}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := val.GetClientCtx().Codec.MarshalJSON(tc.req)
			s.Require().NoError(err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/encode/amino", val.GetAPIAddress()), "application/json", req)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.TxEncodeAminoResponse
				err := val.GetClientCtx().Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)

				var decodedTx legacytx.StdTx
				err = val.GetClientCtx().LegacyAmino.Unmarshal(result.AminoBinary, &decodedTx)
				s.Require().NoError(err)
				s.Require().Equal(decodedTx.GetMsgs(), stdTx.GetMsgs())
			}
		})
	}
}

func (s *E2ETestSuite) readTestAminoTxBinary() ([]byte, *legacytx.StdTx) {
	val := s.network.GetValidators()[0]
	txJSONBytes, err := os.ReadFile("testdata/tx_amino1.bin")
	s.Require().NoError(err)
	var stdTx legacytx.StdTx
	err = val.GetClientCtx().LegacyAmino.Unmarshal(txJSONBytes, &stdTx)
	s.Require().NoError(err)
	return txJSONBytes, &stdTx
}

func (s *E2ETestSuite) TestTxDecodeAmino_GRPC() {
	encodedTx, stdTx := s.readTestAminoTxBinary()

	invalidTxBytes := append(encodedTx, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxDecodeAminoRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: invalidTxBytes}, true, "invalid request"},
		{"valid request with tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: encodedTx}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.queryClient.TxDecodeAmino(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
				s.Require().Empty(res)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(res.GetAminoJson())

				var decodedTx legacytx.StdTx
				err = s.network.GetValidators()[0].GetClientCtx().LegacyAmino.UnmarshalJSON([]byte(res.GetAminoJson()), &decodedTx)
				s.Require().NoError(err)
				s.Require().Equal(stdTx.GetMsgs(), decodedTx.GetMsgs())
			}
		})
	}
}

func (s *E2ETestSuite) TestTxDecodeAmino_GRPCGateway() {
	val := s.network.GetValidators()[0]
	encodedTx, stdTx := s.readTestAminoTxBinary()

	invalidTxBytes := append(encodedTx, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxDecodeAminoRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: invalidTxBytes}, true, "invalid request"},
		{"valid request with tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: encodedTx}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := val.GetClientCtx().Codec.MarshalJSON(tc.req)
			s.Require().NoError(err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/decode/amino", val.GetAPIAddress()), "application/json", req)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.TxDecodeAminoResponse
				err := val.GetClientCtx().Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)

				var decodedTx legacytx.StdTx
				err = val.GetClientCtx().LegacyAmino.UnmarshalJSON([]byte(result.AminoJson), &decodedTx)
				s.Require().NoError(err)
				s.Require().Equal(stdTx.GetMsgs(), decodedTx.GetMsgs())
			}
		})
	}
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
