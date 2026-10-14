package simulation

import (
	"github.com/cosmos/cosmos-sdk/client"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// SimulateMsgSubmitProposalWithVotes is SimulateMsgSubmitProposal with votes
// scheduled for every submitted proposal instead of 1%, so tests can check the
// scheduled votes directly.
func SimulateMsgSubmitProposalWithVotes(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
	msgSim simtypes.MsgSimulatorFn,
) simtypes.Operation {
	return simulateMsgSubmitProposalFromMsgSim(txGen, ak, bk, k, msgSim, 100)
}
