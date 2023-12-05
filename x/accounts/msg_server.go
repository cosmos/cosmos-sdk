package accounts

import (
	"context"

	"cosmossdk.io/x/accounts/internal/implementation"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

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

	// decode message bytes into the concrete boxed message type
	msg, err := unwrapAny(request.Message)
	if err != nil {
		return nil, err
	}

	// run account creation logic
	resp, accAddr, err := m.k.Init(ctx, request.AccountType, creator, msg)
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

	anyResp, err := wrapAny(resp.(proto.Message))
	if err != nil {
		return nil, err
	}
	return &v1.MsgInitResponse{
		AccountAddress: accAddrString,
		Response:       anyResp,
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

	// decode message bytes into the concrete boxed message type
	req, err := unwrapAny(execute.Message)
	if err != nil {
		return nil, err
	}

	// run account execution logic
	resp, err := m.k.Execute(ctx, targetAddr, senderAddr, req)
	if err != nil {
		return nil, err
	}

	// encode the response
	respAny, err := wrapAny(resp.(proto.Message))
	if err != nil {
		return nil, err
	}
	return &v1.MsgExecuteResponse{
		Response: respAny,
	}, nil
}

func (m msgServer) ExecuteBundle(ctx context.Context, req *v1.MsgExecuteBundle) (*v1.MsgExecuteBundleResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func unwrapAny(msg *codectypes.Any) (proto.Message, error) {
	return anypb.UnmarshalNew(&anypb.Any{
		TypeUrl: msg.TypeUrl,
		Value:   msg.Value,
	}, proto.UnmarshalOptions{})
}

func wrapAny(msg proto.Message) (*codectypes.Any, error) {
	anyMsg, err := implementation.PackAny(msg)
	if err != nil {
		return nil, err
	}
	return &codectypes.Any{
		TypeUrl: anyMsg.TypeUrl,
		Value:   anyMsg.Value,
	}, nil
}
