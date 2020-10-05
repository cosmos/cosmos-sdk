package v1

import (
	"context"

	grpc2 "google.golang.org/grpc"

	"github.com/gogo/protobuf/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Module interface {
	Init(Configurator)
}

type Configurator interface {
	StoreKey() StoreKey

	BinaryMarshaler() codec.BinaryMarshaler
	InterfaceRegistry() types.InterfaceRegistry

	MsgServer() grpc.Server
	QueryServer() grpc.Server
	HookServer() grpc.Server

	RequireMsgServices(...interface{})
	RequireQueryServices(...interface{})
}

type StoreKey interface {
	ModuleKey

	KVStore(context.Context) sdk.KVStore
	TransientStore(context.Context) sdk.KVStore
}

type ModuleKey interface {
	grpc2.ClientConnInterface

	DerivedKey(path string) ModuleKey
}

func NewHookClientConn(hookAddr sdk.AccAddress) grpc2.ClientConnInterface {
	panic("TODO")
}
