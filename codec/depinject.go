package codec

import (
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/address"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

func ProvideInterfaceRegistry(
	addressCodec address.Codec,
	validatorAddressCodec address.ValidatorAddressCodec,
	customGetSigners []signing.CustomGetSigner,
) (types.InterfaceRegistry, error) {
	signingOptions := signing.Options{
		AddressCodec:          addressCodec,
		ValidatorAddressCodec: validatorAddressCodec,
	}
	for _, signer := range customGetSigners {
		signingOptions.DefineCustomGetSigners(signer.MsgType, signer.Fn)
	}

	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	if err != nil {
		return nil, err
	}

	if err := interfaceRegistry.SigningContext().Validate(); err != nil {
		return nil, err
	}

	return interfaceRegistry, nil
}
