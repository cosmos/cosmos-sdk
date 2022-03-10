package example

import (
	"context"
	"embed"
	"fmt"
	"github.com/cosmos/cosmos-sdk/core/extension"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/core/app"
	"github.com/cosmos/cosmos-sdk/core/blockinfo"
	"github.com/cosmos/cosmos-sdk/core/event"
	"github.com/cosmos/cosmos-sdk/core/module"
	"github.com/cosmos/cosmos-sdk/core/store"
)

// go:embed proto_image.bin.gz
var pinnedProtoImage embed.FS

func init() {
	// register the module with the app-wiring dependency injection framework
	module.Register(&Module{}, pinnedProtoImage,
		module.Provide(func(inputs Inputs) *app.Handler {
			s := keeper{
				kvStoreKey:       inputs.KVStoreKey,
				blockInfoService: inputs.BlockInfoService,
			}

			inputs.ExtResolver.Register(&msgRegisterNameValidator{})

			h := app.NewHandler()
			RegisterMsgServer(h, s)
			RegisterQueryServer(h, s)
			return h
		}),
	)
}

// the module's dependency injection inputs
type Inputs struct {
	container.In

	KVStoreKey       *store.KVStoreKey
	BlockInfoService blockinfo.Service
	ExtResolver      extension.Resolver
}

// implement ValidateBasic
type msgRegisterNameValidator struct{ *MsgRegisterName }

func (m msgRegisterNameValidator) ValidateBasic() error {
	if m.Sender == "" {
		return fmt.Errorf("missing signer")
	}

	if m.Name == "" {
		return fmt.Errorf("missing signer")
	}

	return nil
}

type keeper struct {
	kvStoreKey       *store.KVStoreKey
	blockInfoService blockinfo.Service
}

const (
	nameInfoPrefix byte = iota
)

func nameInfoKey(name string) []byte {
	return append([]byte{nameInfoPrefix}, name...)
}

// implement MsgServer
func (s keeper) RegisterName(ctx context.Context, msg *MsgRegisterName) (*MsgRegisterNameResponse, error) {
	kvStore := s.kvStoreKey.Open(ctx)
	key := nameInfoKey(msg.Name)
	if kvStore.Has(key) {
		return nil, status.Error(codes.AlreadyExists, "name already registered")
	}

	height := s.blockInfoService.GetBlockInfo(ctx).Height()
	bz, err := proto.Marshal(&NameInfo{
		Owner:            msg.Sender,
		RegisteredHeight: height,
	})
	if err != nil {
		return nil, err
	}

	kvStore.Set(key, bz)
	err = event.GetManager(ctx).Emit(&EventRegisterName{
		Name:  msg.Name,
		Owner: msg.Sender,
	})
	return &MsgRegisterNameResponse{}, err
}

// implement QueryServer
func (s keeper) Name(ctx context.Context, request *QueryNameRequest) (*QueryNameResponse, error) {
	kvStore := s.kvStoreKey.Open(ctx)
	key := nameInfoKey(request.Name)
	bz := kvStore.Get(key)
	if bz == nil {
		return nil, status.Error(codes.NotFound, "name not found")
	}

	var info NameInfo
	err := proto.Unmarshal(bz, &info)
	if err != nil {
		return nil, err
	}

	return &QueryNameResponse{Info: &info}, nil
}

func (s keeper) mustEmbedUnimplementedMsgServer()   {}
func (s keeper) mustEmbedUnimplementedQueryServer() {}

var _ MsgServer = keeper{}
var _ QueryServer = keeper{}
