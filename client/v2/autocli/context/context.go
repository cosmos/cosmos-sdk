package context

import (
	gocontext "context"
	"errors"

	"github.com/spf13/pflag"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/client/v2/autocli/print"
	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/codec"
)

// key is a custom type used as a context key to prevent collisions in the context.Context value store.
type key string

// ContextKey is the key used to store and retrieve the autocli.Context from a context.Context.
const ContextKey key = "autocli.context"

// Context represents the client context used in autocli commands.
// It contains various components needed for command execution.
type Context struct {
	Flags *pflag.FlagSet

	AddressCodec          address.Codec
	ValidatorAddressCodec address.ValidatorAddressCodec
	ConsensusAddressCodec address.ConsensusAddressCodec

	Cdc codec.Codec

	Printer *print.Printer

	Keyring keyring.Keyring

	EnabledSignmodes []apisigning.SignMode
}

// ClientContextFromGoContext returns the autocli.Context from a given Go context.
// It checks if the context contains a valid autocli.Context and returns it.
func ClientContextFromGoContext(ctx gocontext.Context) (*Context, error) {
	if c := ctx.Value(ContextKey); c != nil {
		cliCtx, ok := c.(Context)
		if !ok {
			return nil, errors.New("context value is not of type autocli.Context")
		}
		return &cliCtx, nil
	}
	return nil, errors.New("context does not contain autocli.Context value")
}
