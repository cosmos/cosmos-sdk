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

// ContextKey is the key used to store and retrieve the autocli.Context from a context.Context.
const ContextKey = "autocli.context"

// Context represents the client context used in autocli commands.
// It contains various components needed for command execution.
type Context struct {
	gocontext.Context

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
			return nil, errors.New("invalid context")
		}
		return &cliCtx, nil
	}
	return nil, errors.New("invalid context")
}
