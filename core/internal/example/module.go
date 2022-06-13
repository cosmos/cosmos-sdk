package example

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/blockinfo"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/depinject"
)

func init() {
	// register the module with the app-wiring dependency injection framework
	appmodule.Register(&Module{},
		appmodule.Provide(func(k keeper) *appmodule.Handler {
			h := &appmodule.Handler{}
			RegisterMsgServer(h, k)
			RegisterQueryServer(h, k)
			return h
		}),
	)
}

// the module's dependency injection inputs
type keeper struct {
	depinject.In

	KVStoreKey       store.KVStoreKey
	BlockInfoService blockinfo.Service
	EventService     event.Service
	GasService       gas.Service
}

const (
	nameInfoPrefix byte = iota
)

func nameInfoKey(name string) []byte {
	return append([]byte{nameInfoPrefix}, name...)
}

// implement MsgServer
func (s keeper) RegisterName(ctx context.Context, msg *MsgRegisterName) (*MsgRegisterNameResponse, error) {
	kvStore := s.KVStoreKey.Open(ctx)
	key := nameInfoKey(msg.Name)
	if kvStore.Has(key) {
		return nil, status.Error(codes.AlreadyExists, "name already registered")
	}

	height := s.BlockInfoService.GetBlockInfo(ctx).Height()
	bz, err := proto.Marshal(&NameInfo{
		Owner:            msg.Sender,
		RegisteredHeight: height,
	})
	if err != nil {
		return nil, err
	}

	kvStore.Set(key, bz)
	err = s.EventService.GetManager(ctx).Emit(&EventRegisterName{
		Name:  msg.Name,
		Owner: msg.Sender,
	})
	return &MsgRegisterNameResponse{}, err
}

// implement QueryServer
func (s keeper) Name(ctx context.Context, request *QueryNameRequest) (*QueryNameResponse, error) {
	kvStore := s.KVStoreKey.Open(ctx)
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
