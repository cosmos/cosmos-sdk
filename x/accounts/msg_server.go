package accounts

import (
	"context"
	"fmt"

	accountsv1 "cosmossdk.io/api/cosmos/accounts/v1"
	"cosmossdk.io/x/accounts/tempcore/header"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ accountsv1.MsgServer = msgServerImpl[header.Header]{}

func (a Accounts[H]) MsgServer() accountsv1.MsgServer {
	return msgServerImpl[H]{
		accounts: a,
	}
}

type msgServerImpl[H header.Header] struct {
	accounts Accounts[H]
	accountsv1.UnimplementedMsgServer
}

func (m msgServerImpl[H]) Create(ctx context.Context, create *accountsv1.MsgCreate) (*accountsv1.MsgCreateResponse, error) {
	creator, err := m.accounts.addressCodec.StringToBytes(create.Creator)
	if err != nil {
		return nil, err
	}

	accountImpl, exists := m.accounts.accounts[create.AccountType]
	if !exists {
		return nil, fmt.Errorf("unknown account type %s", create.AccountType)
	}

	msg, err := accountImpl.Schemas.InitMsg.DecodeRequest(create.Message)
	if err != nil {
		return nil, err
	}

	addr, resp, err := m.accounts.Create(ctx, create.AccountType, creator, msg)
	if err != nil {
		return nil, err
	}

	encodedResp, err := accountImpl.Schemas.InitMsg.EncodeResponse(resp)
	if err != nil {
		return nil, err
	}

	addrString, err := m.accounts.addressCodec.BytesToString(addr)
	if err != nil {
		return nil, err
	}

	return &accountsv1.MsgCreateResponse{
		Address:  addrString,
		Response: encodedResp,
	}, nil
}

func (m msgServerImpl[H]) Execute(ctx context.Context, execute *accountsv1.MsgExecute) (*accountsv1.MsgExecuteResponse, error) {
	sender, err := m.accounts.addressCodec.StringToBytes(execute.Sender)
	if err != nil {
		return nil, err
	}

	accAddr, err := m.accounts.addressCodec.StringToBytes(execute.Target)
	if err != nil {
		return nil, err
	}

	accountImpl, err := m.accounts.getAccountImpl(ctx, accAddr)
	if err != nil {
		return nil, err
	}

	msg, err := accountImpl.Schemas.ExecuteMsg.DecodeRequest(execute.Message)
	if err != nil {
		return nil, err
	}

	resp, err := m.accounts.Execute(ctx, sender, accAddr, msg)
	if err != nil {
		return nil, err
	}

	encodedResp, err := accountImpl.Schemas.ExecuteMsg.EncodeResponse(resp)
	if err != nil {
		return nil, err
	}

	return &accountsv1.MsgExecuteResponse{
		Response: encodedResp,
	}, nil
}

func (m msgServerImpl[H]) Migrate(ctx context.Context, migrate *accountsv1.MsgMigrate) (*accountsv1.MsgMigrateResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
