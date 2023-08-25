package accounts

import (
	"testing"

	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestMsgServer_Create(t *testing.T) {
	k, ctx := newKeeper(t, map[string]implementation.Account{
		"test": TestAccount{},
	})

	s := NewMsgServer(k)

	initMsg, err := protojson.Marshal(&emptypb.Empty{})
	require.NoError(t, err)

	resp, err := s.Create(ctx, &v1.MsgCreate{
		Sender:      "sender",
		AccountType: "test",
		Message:     initMsg,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
