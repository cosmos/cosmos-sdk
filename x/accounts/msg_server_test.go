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

func TestMsgServer_UnauthorizedExecution(t *testing.T) {
	k, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	s := NewMsgServer(k)

	// Pack an empty message payload for initialization.
	initMsg, err := implementation.PackAny(&emptypb.Empty{})
	require.NoError(t, err)

	// Initialize a new account with the legitimate owner "owner".
	initResp, err := s.Init(ctx, &v1.MsgInit{
		Sender:      "owner",
		AccountType: "test",
		Message:     initMsg,
		AddressSeed: []byte("seed_owner"),
	})
	require.NoError(t, err)
	require.NotNil(t, initResp)

	// Create the execute message representing a funds transfer (e.g., trying to move "10" funds).
	hackMsg := &wrapperspb.StringValue{
		Value: "10",
	}
	hackMsgAny, err := implementation.PackAny(hackMsg)
	require.NoError(t, err)

	// Hacker attempts to execute a transaction on the legitimate account.
	// Since the "Sender" in the execute message is "hacker" which does NOT
	// match the account owner ("owner"), the execution should fail.
	_, err = s.Execute(ctx, &v1.MsgExecute{
		Sender:  "hacker", // attacker trying to control the funds
		Target:  initResp.AccountAddress,
		Message: hackMsgAny,
	})
	require.Error(t, err, "expected error when unauthorized hacker tries to execute transaction")
}
