package appmodule

import (
	"context"
)

// HasGenesis defines a custom genesis handling API implementation.
// TODO: finalize this API
type HasGenesis interface {
	AppModule
	DefaultGenesis() Message
	ValidateGenesis(data Message) error
	InitGenesis(ctx context.Context, data Message) error
	ExportGenesis(ctx context.Context) (Message, error)
}
