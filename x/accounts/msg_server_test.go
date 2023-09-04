package accounts

import (
	"testing"

	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestMsgServer(t *testing.T) {
	k, ctx := newKeeper(t, map[string]implementation.Account{
		"test": TestAccount{},
	})

	s := NewMsgServer(k)

	// create
	initMsg, err := proto.Marshal(&emptypb.Empty{})
	require.NoError(t, err)

	initResp, err := s.Create(ctx, &v1.MsgCreate{
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
