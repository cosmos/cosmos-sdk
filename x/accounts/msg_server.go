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
