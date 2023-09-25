package autocli

import (
	"errors"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/log"
)

// Builder manages options for building CLI commands.
type Builder struct {
	// flag.Builder embeds the flag builder and its options.
	flag.Builder

	// Logger is the logger used by the builder.
	Logger log.Logger

	// GetClientConn specifies how CLI commands will resolve a grpc.ClientConnInterface
	// from a given context.
	GetClientConn func(*cobra.Command) (grpc.ClientConnInterface, error)

	// AddQueryConnFlags and AddTxConnFlags are functions that add flags to query and transaction commands
	AddQueryConnFlags func(*cobra.Command)
	AddTxConnFlags    func(*cobra.Command)
}

// ValidateAndComplete the builder fields.
// It returns an error if any of the required fields are missing.
// If the Logger is nil, it will be set to a nop logger.
// If the keyring is nil, it will be set to a no keyring.
func (b *Builder) ValidateAndComplete() error {
	if b.Logger == nil {
		b.Logger = log.NewNopLogger()
	}

	if b.ClientCtx == nil {
		return errors.New("client context is required in builder")
	}

	if b.ClientCtx.AddressCodec == nil {
		return errors.New("address codec is required in builder")
	}

	if b.ClientCtx.ValidatorAddressCodec == nil {
		return errors.New("validator address codec is required in builder")
	}

	if b.ClientCtx.ConsensusAddressCodec == nil {
		return errors.New("consensus address codec is required in builder")
	}

	if b.Keyring == nil {
		if b.ClientCtx.Keyring != nil {
			b.Keyring = b.ClientCtx.Keyring
		} else {
			b.Keyring = keyring.NoKeyring{}
		}
	}

	if b.TypeResolver == nil {
		return errors.New("type resolver is required in builder")
	}

	if b.FileResolver == nil {
		return errors.New("file resolver is required in builder")
	}

	return nil
}
