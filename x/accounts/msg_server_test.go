package accounts

import (
	"testing"

	"cosmossdk.io/x/accounts/accountstd"
	"github.com/stretchr/testify/require"

	v1 "cosmossdk.io/x/accounts/v1"
)

func TestMsgServer(t *testing.T) {
	k, ctx := newKeeper(t, accountstd.AddAccount("test", NewTestAccount))
	s := NewMsgServer(k)

	// create
	initResp, err := s.Init(ctx, &v1.MsgInit{
		Sender:      "sender",
		AccountType: "test",
		JsonMessage: `{}`,
	})
	require.NoError(t, err)
	require.NotNil(t, initResp)

	// execute
	execResp, err := s.Execute(ctx, &v1.MsgExecute{
		Sender:            "sender",
		Target:            initResp.AccountAddress,
		ExecuteMsgTypeUrl: "google.protobuf.UInt64Value",
		JsonMessage:       `10`,
	})
	require.NoError(t, err)
	require.NotNil(t, execResp)
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
