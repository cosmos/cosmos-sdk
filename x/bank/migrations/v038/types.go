// Package v038 is used for legacy migration scripts. Actual migration scripts
// for v038 have been removed, but the v039->v042 migration script still
// references types from this file, so we're keeping it for now.
package v038

// DONTCOVER

const (
	ModuleName = "bank"
)

type (
	GenesisState struct {
		SendEnabled bool `json:"send_enabled" yaml:"send_enabled"`
	}
)
