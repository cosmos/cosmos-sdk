package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"cosmossdk.io/x/epochs/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// GenCommunityTax randomized CommunityTax
func GenDuration(r *rand.Rand) time.Duration {
	return time.Hour * time.Duration(r.Intn(168)+1) // between 1 hour to 1 week
}

func RandomizedEpochs(r *rand.Rand) []types.EpochInfo {
	// Gen max 10 epoch
	n := r.Intn(11)
	var epochs []types.EpochInfo
	for i := 0; i < n; i++ {
		identifier := "identifier-" + strconv.Itoa(i)
		duration := GenDuration(r)
		epoch := types.NewGenesisEpochInfo(identifier, duration)
		epochs = append(epochs, epoch)
	}
	return epochs
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	epochs := RandomizedEpochs(simState.Rand)
	epochsGenesis := types.GenesisState{
		Epochs: epochs,
	}

	bz, err := json.MarshalIndent(&epochsGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated epochs parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&epochsGenesis)
}
