package autocli

import (
	"github.com/spf13/cobra"

	"cosmossdk.io/client/v2/autocli/flag"
	"github.com/cosmos/cosmos-sdk/codec"
)

// Builder manages options for building CLI commands.
type Builder struct {
	// flag.Builder embeds the flag builder and its options.
	flag.Builder

	// AddQueryConnFlags and AddTxConnFlags are functions that add flags to query and transaction commands
	AddQueryConnFlags func(*cobra.Command)
	AddTxConnFlags    func(*cobra.Command)

	Cdc codec.Codec
}

// ValidateAndComplete the builder fields.
// It returns an error if any of the required fields are missing.
func (b *Builder) ValidateAndComplete() error {
	return b.Builder.ValidateAndComplete()
}
