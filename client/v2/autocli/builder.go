package autocli

import (
	"errors"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/client/v2/autocli/keyring"
)

// Builder manages options for building CLI commands.
type Builder struct {
	// flag.Builder embeds the flag builder and its options.
	flag.Builder

	// GetClientConn specifies how CLI commands will resolve a grpc.ClientConnInterface
	// from a given context.
	GetClientConn func(*cobra.Command) (grpc.ClientConnInterface, error)

	AddQueryConnFlags func(*cobra.Command)

	AddTxConnFlags func(*cobra.Command)
}

func (b *Builder) Validate() error {
	if b.AddressCodec == nil {
		return errors.New("address codec is required in builder")
	}

	if b.ValidatorAddressCodec == nil {
		return errors.New("validator address codec is required in builder")
	}

	if b.ConsensusAddressCodec == nil {
		return errors.New("consensus address codec is required in builder")
	}

	if b.TypeResolver == nil {
		return errors.New("type resolver is required in builder")
	}

	if b.FileResolver == nil {
		return errors.New("file resolver is required in builder")
	}

	if b.Keyring == nil {
		b.Keyring = keyring.NoKeyring{}
	}

	return nil
}
