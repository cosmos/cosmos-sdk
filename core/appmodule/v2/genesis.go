package appmodule

import (
	"context"
)

// HasGenesis defines a custom genesis handling API implementation.
// WARNING: this API is meant as a short-term solution to allow for the
// migration of existing modules to the new app module API. It is intended to be replaced by collections
type HasGenesis interface {
	AppModule
	DefaultGenesis() Message
	ValidateGenesis(data Message) error
	InitGenesis(ctx context.Context, data Message) error
	ExportGenesis(ctx context.Context) (Message, error)
}
