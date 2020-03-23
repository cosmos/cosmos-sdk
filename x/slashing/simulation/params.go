package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

const (
	keySignedBlocksWindow    = "SignedBlocksWindow"
	keyMinSignedPerWindow    = "MinSignedPerWindow"
	keySlashFractionDowntime = "SlashFractionDowntime"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keySignedBlocksWindow,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenSignedBlocksWindow(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyMinSignedPerWindow,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMinSignedPerWindow(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keySlashFractionDowntime,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenSlashFractionDowntime(r))
			},
		),
	}
}
