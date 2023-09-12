package accounts

import (
	"context"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
)

func TestMsgServer(t *testing.T) {
	k, ctx := newKeeper(t, map[string]implementation.Account{
		"test": TestAccount{},
	})
	k.queryModuleFunc = func(ctx context.Context, msg proto.Message) (proto.Message, error) {
		_, ok := msg.(*bankv1beta1.QueryBalanceRequest)
		require.True(t, ok)
		return &bankv1beta1.QueryBalanceResponse{}, nil
	}

	s := NewMsgServer(k)

	// create
	initMsg, err := proto.Marshal(&emptypb.Empty{})
	require.NoError(t, err)

	initResp, err := s.Init(ctx, &v1.MsgInit{
		Sender:      "sender",
		AccountType: "test",
		Message:     initMsg,
	})
	require.NoError(t, err)
	require.NotNil(t, initResp)

	// execute
	executeMsg := &wrapperspb.StringValue{
		Value: "10",
	}
	executeMsgAny, err := anypb.New(executeMsg)
	require.NoError(t, err)

	executeMsgBytes, err := proto.Marshal(executeMsgAny)
	require.NoError(t, err)

	execResp, err := s.Execute(ctx, &v1.MsgExecute{
		Sender:  "sender",
		Target:  initResp.AccountAddress,
		Message: executeMsgBytes,
	})
	require.NoError(t, err)
	require.NotNil(t, execResp)
}
