package offchain

import (
	"context"
	"errors"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
)

const mnemonic = "have embark stumble card pistol fun gauge obtain forget oil awesome lottery unfold corn sure original exist siren pudding spread uphold dwarf goddess card"

func getCodec() codec.Codec {
	registry := testutil.CodecOptions{}.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)

	return codec.NewProtoCodec(registry)
}

var _ gogogrpc.ClientConn = mockClientConn{}

type mockClientConn struct{}

func (c mockClientConn) Invoke(_ context.Context, _ string, _, _ interface{}, _ ...grpc.CallOption) error {
	return errors.New("not implemented")
}

func (c mockClientConn) NewStream(_ context.Context, _ *grpc.StreamDesc, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("not implemented")
}
