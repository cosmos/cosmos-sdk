package types

import (
	context "context"

	"cosmossdk.io/core/address"
	"google.golang.org/protobuf/runtime/protoiface"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProtoMsg = protoiface.MessageV1

// AuthKeeper defines the auth contract that must be fulfilled when
// creating a x/accountlink keeper.
type AuthKeeper interface {
	AddressCodec() address.Codec

	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
}

// AccountKeeper defines the account contract that must be fulfilled when
// creating a x/accountlink keeper.
type AccountKeeper interface {
	Execute(
		ctx context.Context,
		accountAddr []byte,
		sender []byte,
		execRequest ProtoMsg,
	) (ProtoMsg, error)
}
