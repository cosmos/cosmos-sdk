// This file contains legacy code for testing `ServiceMsg`s. Following the issue
// https://github.com/cosmos/cosmos-sdk/issues/9063, `ServiceMsg`s have been
// deprecated, but we still support them in v043. These tests ensures that txs
// containing `ServiceMsg`s still work.
//
// This file should be removed as part of https://github.com/cosmos/cosmos-sdk/issues/9172.
package legacytx_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MsgRequest is the interface a transaction message, defined as a proto
// service method, must fulfill.
type msgRequest interface {
	proto.Message
	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error
	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []sdk.AccAddress
}

// ServiceMsg is the struct into which an Any whose typeUrl matches a service
// method format (ex. `/cosmos.gov.v1beta1.Msg/SubmitProposal`) unpacks.
type serviceMsg struct {
	// MethodName is the fully-qualified service method name.
	MethodName string
	// Request is the request payload.
	Request msgRequest
}

var _ sdk.Msg = &serviceMsg{}

func (msg *serviceMsg) ProtoMessage()  {}
func (msg *serviceMsg) Reset()         {}
func (msg *serviceMsg) String() string { return "ServiceMsg" }

// ValidateBasic implements Msg.ValidateBasic method.
func (msg *serviceMsg) ValidateBasic() error {
	return msg.Request.ValidateBasic()
}

// GetSigners implements Msg.GetSigners method.
func (msg *serviceMsg) GetSigners() []sdk.AccAddress {
	return msg.Request.GetSigners()
}

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)
	s.Require().NotNil(s.network)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// TestServiceMsg tests sending txs with a ServiceMsg.
func (s IntegrationTestSuite) TestServiceMsg() {
	val := s.network.Validators[0]

	// prepare txBuilder with msg
	txBuilder := val.ClientCtx.TxConfig.NewTxBuilder()

	// This sets a ServiceMsg Msg/Send.
	// Thanks for the `proto.Marshal` override at the end of this file, tx
	// encoding works nicely.
	err := txBuilder.SetMsgs(&serviceMsg{
		MethodName: "/cosmos.bank.v1beta1.Msg/Send",
		Request: &types.MsgSend{
			FromAddress: val.Address.String(),
			ToAddress:   val.Address.String(),
			Amount:      sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 23)},
		},
	})
	s.Require().NoError(err)

	txBuilder.SetFeeAmount(sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 23)})
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())

	// setup txFactory
	txFactory := tx.Factory{}.
		WithChainID(val.ClientCtx.ChainID).
		WithKeybase(val.ClientCtx.Keyring).
		WithTxConfig(val.ClientCtx.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	// Sign Tx.
	err = authclient.SignTx(txFactory, val.ClientCtx, val.Moniker, txBuilder, false, true)
	s.Require().NoError(err)

	txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	// Broadcast the tx via gRPC.
	queryClient := txtypes.NewServiceClient(val.ClientCtx)

	grpcRes, err := queryClient.BroadcastTx(
		context.Background(),
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), grpcRes.TxResponse.Code)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// newMarshaler is the interface representing objects that can marshal themselves.
// This exists to overwrite `proto.Marshal` behavior for ServiceMsg.
//
// DO NOT DEPEND ON THIS.
type svcMsgMarshaler interface {
	XXX_Size() int
	XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
}

// Overwrite the `proto.MessageName` and `proto.Marshal` return values for ServiceMsg.
func (msg *serviceMsg) XXX_MessageName() string { return strings.TrimPrefix(msg.MethodName, "/") }
func (msg *serviceMsg) XXX_Size() int           { return msg.Request.(svcMsgMarshaler).XXX_Size() }
func (msg *serviceMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return msg.Request.(svcMsgMarshaler).XXX_Marshal(b, deterministic)
}
