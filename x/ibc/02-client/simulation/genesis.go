package simulation

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"
	"math/rand"
	"time"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// GenClientGenesis returns the default client genesis state.
func GenClientGenesis(_ *rand.Rand, _ []simtypes.Account) types.GenesisState {
	//return types.DefaultGenesisState()

	var (
		clientID = "ethbridge"

		trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
		ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
		maxClockDrift  time.Duration = time.Second * 10
	)

	privVal := tmtypes.NewMockPV()
	pubKey, err := privVal.GetPubKey()
	if err !=nil {
		panic(err)
	}

	now := time.Now().UTC()

	val := tmtypes.NewValidator(pubKey, 10)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{val})

	header := ibctmtypes.CreateTestHeader("chain_B", 10, now, valSet, []tmtypes.PrivValidator{privVal})

	return types.NewGenesisState(
		[]exported.ClientState{
			ibctmtypes.NewClientState(clientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, header, commitmenttypes.GetSDKSpecs()),
			localhosttypes.NewClientState("chain_B", 10),
		},
		[]types.ClientConsensusStates{
			{
				clientID,
				[]exported.ConsensusState{
					ibctmtypes.NewConsensusState(
						header.Time, commitmenttypes.NewMerkleRoot(header.AppHash), header.GetHeight(), header.ValidatorSet,
					),
				},
			},
		},
		true,
	)
}
