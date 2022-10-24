package rosetta

import (
	"github.com/pointnetwork/cosmos-point-sdk/codec"
	codectypes "github.com/pointnetwork/cosmos-point-sdk/codec/types"
	cryptocodec "github.com/pointnetwork/cosmos-point-sdk/crypto/codec"
	authcodec "github.com/pointnetwork/cosmos-point-sdk/x/auth/types"
	bankcodec "github.com/pointnetwork/cosmos-point-sdk/x/bank/types"
)

// MakeCodec generates the codec required to interact
// with the cosmos APIs used by the rosetta gateway
func MakeCodec() (*codec.ProtoCodec, codectypes.InterfaceRegistry) {
	ir := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)

	authcodec.RegisterInterfaces(ir)
	bankcodec.RegisterInterfaces(ir)
	cryptocodec.RegisterInterfaces(ir)

	return cdc, ir
}
