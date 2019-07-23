package types

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"
)

// DONTCOVER

type (
	// AppMap map modules names with their json raw representation
	AppMap map[string]json.RawMessage
	// MigrationCallback converts a genesis map from the previous version to the targeted one
	MigrationCallback func(AppMap) AppMap
	// MigrationMap defines a mapping from a version to a MigrationCallback
	MigrationMap map[string]MigrationCallback
)

// ModuleName is genutil
const ModuleName = "genutil"

// InitConfig common config options for init
type InitConfig struct {
	ChainID   string
	GenTxsDir string
	Name      string
	NodeID    string
	ValPubKey crypto.PubKey
}

// NewInitConfig creates a new InitConfig object
func NewInitConfig(chainID, genTxsDir, name, nodeID string, valPubKey crypto.PubKey) InitConfig {
	return InitConfig{
		ChainID:   chainID,
		GenTxsDir: genTxsDir,
		Name:      name,
		NodeID:    nodeID,
		ValPubKey: valPubKey,
	}
}
