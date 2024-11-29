package context

import (
	gocontext "context"
	"errors"

	"github.com/spf13/pflag"

	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/codec"
)

// ContextKey is a key used to store and retrieve Context from a Go context.Context.
var ContextKey contextKey

// contextKey is an empty struct used as a key type for storing Context in a context.Context.
type contextKey struct{}

// Context represents the client context used in autocli commands.
// It contains various components needed for command execution.
type Context struct {
	Flags *pflag.FlagSet

	AddressCodec          address.Codec
	ValidatorAddressCodec address.ValidatorAddressCodec
	ConsensusAddressCodec address.ConsensusAddressCodec

	Cdc codec.Codec

	Keyring keyring.Keyring

	EnabledSignModes []apisigning.SignMode
}

// SetInContext stores the provided autocli.Context in the given Go context.Context.
// It returns a new context.Context containing the autocli.Context value.
func SetInContext(goCtx gocontext.Context, cliCtx Context) gocontext.Context {
	return gocontext.WithValue(goCtx, ContextKey, cliCtx)
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
