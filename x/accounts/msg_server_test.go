package accounts

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
)

func TestMsgServer(t *testing.T) {
	k, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	s := NewMsgServer(k)

	// create
	initMsg, err := implementation.PackAny(&emptypb.Empty{})
	require.NoError(t, err)

	initResp, err := s.Init(ctx, &v1.MsgInit{
		Sender:      "sender",
		AccountType: "test",
		Message:     initMsg,
		AddressSeed: []byte("seed"),
	})
	require.NoError(t, err)
	require.NotNil(t, initResp)

	t.Run("success", func(t *testing.T) {
		// execute
		executeMsg := &wrapperspb.StringValue{
			Value: "10",
		}
		executeMsgAny, err := implementation.PackAny(executeMsg)
		require.NoError(t, err)

		execResp, err := s.Execute(ctx, &v1.MsgExecute{
			Sender:  "sender",
			Target:  initResp.AccountAddress,
			Message: executeMsgAny,
		})
		require.NoError(t, err)
		require.NotNil(t, execResp)
	})

	t.Run("fail initting same account twice", func(t *testing.T) {
		_, err := s.Init(ctx, &v1.MsgInit{
			Sender:      "sender",
			AccountType: "test",
			Message:     initMsg,
			AddressSeed: []byte("seed"),
		})
		require.ErrorIs(t, err, ErrAccountAlreadyExists)
	})

	t.Run("initting without seed", func(t *testing.T) {
		_, err := s.Init(ctx, &v1.MsgInit{
			Sender:      "sender",
			AccountType: "test",
			Message:     initMsg,
		})
		require.NoError(t, err)
	})
}

func TestMsgServer_BundlingDisabled(t *testing.T) {
	k, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	k.DisableTxBundling()

	s := NewMsgServer(k)

	_, err := s.ExecuteBundle(ctx, &v1.MsgExecuteBundle{
		Bundler: "someone",
		Txs:     nil,
	})
	require.ErrorIs(t, err, ErrBundlingDisabled)
}
