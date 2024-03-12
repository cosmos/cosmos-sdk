package appmodule

import (
	"context"
	"encoding/json"
)

// HasGenesis defines a custom genesis handling API implementation.
// WARNING: this API is meant as a short-term solution to allow for the
// migration of existing modules to the new app module API. It is intended to be replaced by collections
type HasGenesis interface {
	AppModule
	DefaultGenesis() json.RawMessage
	ValidateGenesis(data json.RawMessage) error
	InitGenesis(ctx context.Context, data json.RawMessage) error
	ExportGenesis(ctx context.Context) (json.RawMessage, error)
}
