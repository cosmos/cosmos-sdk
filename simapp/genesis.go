package simapp

import (
	"github.com/gogo/protobuf/proto"
)

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() map[string]proto.Message {
	return ModuleBasics.DefaultGenesis()
}
