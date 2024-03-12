package runtime

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	gogotest "github.com/cosmos/cosmos-sdk/testutil/testdata"
	protov2test "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	v1Handler := func(ctx context.Context, msg v1Message, resp v1Message) error {
		m, ok := msg.(*gogotest.MsgCreateDog)
		require.True(t, ok)
		*(resp.(*gogotest.MsgCreateDogResponse)) = gogotest.MsgCreateDogResponse{Name: m.Owner}
		return nil
	}

	_ = func(ctx context.Context, msg v2Message, resp v2Message) error {
		m, ok := msg.(*protov2test.MsgCreateDog)
		require.True(t, ok)
		*(resp.(*protov2test.MsgCreateDogResponse)) = protov2test.MsgCreateDogResponse{
			Name: m.Owner + "v2",
		}
		return nil
	}

	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	preV1Called := false
	preV1 := func(ctx context.Context, msg v1Message) error {
		preV1Called = true
		_, ok := msg.(*gogotest.MsgCreateDog)
		require.True(t, ok)
		return nil
	}

	preV2Called := false
	preV2 := func(ctx context.Context, msg v2Message) error {
		preV2Called = true
		_, ok := msg.(*protov2test.MsgCreateDog)
		require.True(t, ok)
		return nil
	}

	postV1Called := false
	postV1 := func(ctx context.Context, msg v1Message, resp v1Message) error {
		postV1Called = true
		_, ok := msg.(*gogotest.MsgCreateDog)
		require.True(t, ok)
		_, ok = resp.(*gogotest.MsgCreateDogResponse)
		require.True(t, ok)
		return nil
	}

	postV2Called := false
	postV2 := func(ctx context.Context, msg v2Message, resp v2Message) error {
		postV2Called = true
		_, ok := msg.(*protov2test.MsgCreateDog)
		require.True(t, ok)
		_, ok = resp.(*protov2test.MsgCreateDogResponse)
		require.True(t, ok)
		return nil
	}

	h := &hybridHandler{
		cdc:           cdc,
		makeReqV1:     func() v1Message { return new(gogotest.MsgCreateDog) },
		makeRespV1:    func() v1Message { return new(gogotest.MsgCreateDogResponse) },
		makeReqV2:     func() v2Message { return new(protov2test.MsgCreateDog) },
		makeRespV2:    func() v2Message { return new(protov2test.MsgCreateDogResponse) },
		preHandlerV1:  preV1,
		preHandlerV2:  preV2,
		handlerV1:     v1Handler,
		handlerV2:     nil,
		postHandlerV1: postV1,
		postHandlerV2: postV2,
	}

	gogoResp := new(gogotest.MsgCreateDogResponse)

	err := h.handle(context.Background(), &gogotest.MsgCreateDog{
		Dog:   nil,
		Owner: "hello",
	}, gogoResp)
	require.NoError(t, err)
	require.Equal(t, gogoResp.Name, "hello")

	require.True(t, preV1Called)
	require.True(t, preV2Called)
	require.True(t, postV1Called)
	require.True(t, postV2Called)
}
