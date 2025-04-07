package types

import (
	"github.com/cosmos/gogoproto/proto"
)

// LegacyAccountI is the interface that legacy account types must implement to be migrated
// to the new simplified AccountI interface.
type LegacyAccountI interface {
	proto.Message
	Migrate() AccountMigrationData
}

type AccountMigrationData struct {
	Address       []byte
	AccountNumber uint64
	Sequence      uint64
	NewAccount    AccountI
}

type AccountI interface {
	proto.Message
}

// ModuleAccountI defines an account interface for modules that hold tokens in
// an escrow.
type ModuleAccountI interface {
	GetName() string
	GetPermissions() []string
	HasPermission(string) bool
}
