package accounts

import (
	"context"

	v1 "cosmossdk.io/x/accounts/v1"
)

var _ v1.MsgServer = msgServer{}

func NewMsgServer(k Keeper) v1.MsgServer {
	return &msgServer{k}
}

type msgServer struct {
	k Keeper
}

func (m msgServer) Create(ctx context.Context, request *v1.MsgCreate) (*v1.MsgCreateResponse, error) {
	creator, err := m.k.addressCodec.StringToBytes(request.Sender)
	if err != nil {
		return nil, err
	}

	impl, err := m.k.getImplementation(request.AccountType)
	if err != nil {
		return nil, err
	}

	// decode message bytes into the concrete boxed message type
	msg, err := impl.DecodeInitRequest(request.Message)
	if err != nil {
		return nil, err
	}

	// run account creation logic
	resp, accAddr, err := m.k.Create(ctx, request.AccountType, creator, msg)
	if err != nil {
		return nil, err
	}

	// encode the response
	respBytes, err := impl.EncodeInitResponse(resp)
	if err != nil {
		return nil, err
	}

	// encode the address
	accAddrString, err := m.k.addressCodec.BytesToString(accAddr)
	if err != nil {
		return nil, err
	}

	return &v1.MsgCreateResponse{
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
