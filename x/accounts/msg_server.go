package accounts

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/core/event"
	v1 "cosmossdk.io/x/accounts/v1"
)

var _ v1.MsgServer = msgServer{}

func NewMsgServer(k Keeper) v1.MsgServer {
	return &msgServer{k}
}

type msgServer struct {
	k Keeper
}

func (m msgServer) Init(ctx context.Context, request *v1.MsgInit) (*v1.MsgInitResponse, error) {
	creator, err := m.k.addressCodec.StringToBytes(request.Sender)
	if err != nil {
		return nil, err
	}

	impl, err := m.k.getImplementation(request.AccountType)
	if err != nil {
		return nil, err
	}

	// decode message bytes into the concrete boxed message type
	msg, err := impl.InitHandlerSchema.RequestSchema.TxDecode(request.Message)
	if err != nil {
		return nil, err
	}

	// run account creation logic
	resp, accAddr, err := m.k.Init(ctx, request.AccountType, creator, msg)
	if err != nil {
		return nil, err
	}

	// encode the response
	respBytes, err := impl.InitHandlerSchema.ResponseSchema.TxEncode(resp)
	if err != nil {
		return nil, err
	}

	// encode the address
	accAddrString, err := m.k.addressCodec.BytesToString(accAddr)
	if err != nil {
		return nil, err
	}

	eventManager := m.k.eventService.EventManager(ctx)
	err = eventManager.EmitKV(
		ctx,
		"account_creation",
		event.Attribute{
			Key:   "address",
			Value: accAddrString,
		},
	)
	if err != nil {
		return nil, err
	}
	return &v1.MsgInitResponse{
		AccountAddress: accAddrString,
		Response:       respBytes,
	}, nil
}

func (m msgServer) Execute(ctx context.Context, execute *v1.MsgExecute) (*v1.MsgExecuteResponse, error) {
	// decode sender address
	senderAddr, err := m.k.addressCodec.StringToBytes(execute.Sender)
	if err != nil {
		return nil, err
	}
	// decode the target address
	targetAddr, err := m.k.addressCodec.StringToBytes(execute.Target)
	if err != nil {
		return nil, err
	}

	// get account type
	accType, err := m.k.AccountsByType.Get(ctx, targetAddr)
	if err != nil {
		return nil, err
	}

	// get the implementation
	impl, err := m.k.getImplementation(accType)
	if err != nil {
		return nil, err
	}

	// decode message bytes into the concrete boxed message type
	req, err := impl.DecodeExecuteRequest(execute.Message)
	if err != nil {
		return nil, err
	}

	// run account execution logic
	resp, err := m.k.Execute(ctx, targetAddr, senderAddr, req)
	if err != nil {
		return nil, err
	}

	// encode the response
	respBytes, err := impl.EncodeExecuteResponse(resp)
	if err != nil {
		return nil, err
	}

	return &v1.MsgExecuteResponse{
		Response: respBytes,
	}, nil
}

func (m msgServer) ExecuteBundle(ctx context.Context, req *v1.MsgExecuteBundle) (*v1.MsgExecuteBundleResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
