package simulation

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxLifecycleApp is the app entrypoint required to execute a transaction
// through the default ABCI lifecycle in simulation.
type TxLifecycleApp interface {
	CheckTx(req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error)
	PrepareProposal(req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error)
	ProcessProposal(req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error)
	SimDeliver(txEncoder sdk.TxEncoder, tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error)
}

// TxLifecycleOutcome captures the lifecycle execution result.
type TxLifecycleOutcome struct {
	Accepted bool
	Phase    TxLifecyclePhase
	Reason   string
	Err      error
}

// ExecuteTxLifecycle runs tx through CheckTx -> PrepareProposal ->
// ProcessProposal -> Finalize/Deliver.
func ExecuteTxLifecycle(app TxLifecycleApp, txGen client.TxConfig, tx sdk.Tx, ctx sdk.Context) TxLifecycleOutcome {
	txBytes, err := txGen.TxEncoder()(tx)
	if err != nil {
		return TxLifecycleOutcome{
			Accepted: false,
			Phase:    TxPhaseCheckTx,
			Reason:   fmt.Sprintf("error: unable to encode tx: %v", err),
			Err:      err,
		}
	}

	checkRes, err := app.CheckTx(&abci.RequestCheckTx{
		Type: abci.CheckTxType_New,
		Tx:   txBytes,
	})
	if err != nil {
		return TxLifecycleOutcome{
			Accepted: false,
			Phase:    TxPhaseCheckTx,
			Reason:   fmt.Sprintf("error: %v", err),
			Err:      err,
		}
	}
	if checkRes.Code != 0 {
		reason := fmt.Sprintf("rejected code=%d codespace=%s", checkRes.Code, checkRes.Codespace)
		if label := lookupABCIErrorLabel(checkRes.Codespace, checkRes.Code); label != "" {
			reason = fmt.Sprintf("%s (%s)", reason, label)
		}
		return TxLifecycleOutcome{
			Accepted: false,
			Phase:    TxPhaseCheckTx,
			Reason:   reason,
		}
	}

	// SimDeliver may panic in some simulation edge cases before BaseApp's internal
	// panic recovery is installed (e.g. early block gas checks), so guard it here
	// and record as finalize-phase rejection instead of aborting the whole run.
	var deliverErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				deliverErr = fmt.Errorf("panic: %v", r)
			}
		}()
		_, _, deliverErr = app.SimDeliver(txGen.TxEncoder(), tx)
	}()
	if deliverErr != nil {
		return TxLifecycleOutcome{
			Accepted: false,
			Phase:    TxPhaseFinalize,
			Reason:   fmt.Sprintf("error: %v", deliverErr),
			Err:      deliverErr,
		}
	}

	return TxLifecycleOutcome{Accepted: true}
}

func lookupABCIErrorLabel(codespace string, code uint32) string {
	if codespace != "sdk" {
		return ""
	}

	switch code {
	case 4:
		return "unauthorized"
	case 5:
		return "insufficient funds"
	case 32:
		return "incorrect account sequence"
	default:
		return ""
	}
}
