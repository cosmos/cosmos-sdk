package simulate

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type simulateServer struct {
	app         baseapp.BaseApp
	pubkeyCodec cryptotypes.PublicKeyCodec
	txConfig    client.TxConfig
}

// NewSimulateServer creates a new SimulateServer.
func NewSimulateServer(app baseapp.BaseApp, pubkeyCodec cryptotypes.PublicKeyCodec, txConfig client.TxConfig) SimulateServiceServer {
	return simulateServer{
		app:         app,
		pubkeyCodec: pubkeyCodec,
		txConfig:    txConfig,
	}
}

var _ SimulateServiceServer = simulateServer{}

// Simulate implements the SimulateService.Simulate RPC method.
func (s simulateServer) Simulate(ctx context.Context, req *SimulateRequest) (*SimulateResponse, error) {
	if req.Tx == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	err := req.Tx.UnpackInterfaces(s.app.GRPCQueryRouter().InterfaceRegistry())
	if err != nil {
		return nil, err
	}
	txBuilder, err := txBuilderFromProto(s.txConfig, s.pubkeyCodec, req.Tx)
	if err != nil {
		return nil, err
	}

	txBytes, err := req.Tx.Marshal()
	if err != nil {
		return nil, err
	}

	gasInfo, result, err := s.app.Simulate(txBytes, txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return &SimulateResponse{
		GasInfo: &gasInfo,
		Result:  result,
	}, nil
}

// txBuilderFromProto converts a proto.Message Tx into a TxBuilder.
func txBuilderFromProto(txConfig client.TxConfig, pubkeyCodec cryptotypes.PublicKeyCodec, tx *txtypes.Tx) (client.TxBuilder, error) {
	txBuilder := txConfig.NewTxBuilder()

	// Add messages.
	msgs := make([]sdk.Msg, len(tx.Body.Messages))
	for i, any := range tx.Body.Messages {
		msgs[i] = any.GetCachedValue().(sdk.Msg)
	}
	txBuilder.SetMsgs(msgs...)

	// Add other stuff.
	txBuilder.SetMemo(tx.Body.Memo)
	txBuilder.SetFeeAmount(tx.AuthInfo.Fee.Amount)
	txBuilder.SetGasLimit(tx.AuthInfo.Fee.GasLimit)
	txBuilder.SetTimeoutHeight(tx.Body.TimeoutHeight)

	// Add signatures.
	sigs := make([]signing.SignatureV2, len(tx.AuthInfo.SignerInfos))
	for i, signerInfo := range tx.AuthInfo.SignerInfos {
		modeInfo := signerInfo.ModeInfo
		sigData, err := authtx.ModeInfoAndSigToSignatureData(modeInfo, tx.Signatures[i])
		if err != nil {
			return nil, err
		}
		pubKey, err := pubkeyCodec.Decode(signerInfo.PublicKey)
		if err != nil {
			return nil, err
		}

		sigs[i] = signing.SignatureV2{
			PubKey: pubKey,
			Data:   sigData,
		}
	}
	txBuilder.SetSignatures(sigs...)

	return txBuilder, nil
}
