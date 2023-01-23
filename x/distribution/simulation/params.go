package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

const (
	keyCommunityTax            = "communitytax"
	keyBaseProposerReward      = "baseproposerreward"
	keyBonusProposerReward     = "bonusproposerreward"
	keySecretFoundationTax     = "secretfoundationtax"
	keyMinimumRestakeThreshold = "minimumrestakethreshold"
	keyRestakePeriod           = "restakeperiod"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyCommunityTax,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenCommunityTax(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keySecretFoundationTax,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenSecretFoundationTax(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyBaseProposerReward,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBaseProposerReward(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyBonusProposerReward,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBonusProposerReward(r))
			},
		),

		simulation.NewSimParamChange(types.ModuleName, keyMinimumRestakeThreshold,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMinimumRestakeThreshold(r))
			},
		),

		simulation.NewSimParamChange(types.ModuleName, keyRestakePeriod,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenRestakePeriod(r))
			},
		),
	}
}
