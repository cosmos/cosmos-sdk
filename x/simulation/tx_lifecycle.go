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

	maxTxBytes := int64(len(txBytes) + 1024)
	prepareRes, err := app.PrepareProposal(&abci.RequestPrepareProposal{
		Height:     ctx.BlockHeight(),
		Time:       ctx.BlockTime(),
		MaxTxBytes: maxTxBytes,
	})
	if err != nil {
		return TxLifecycleOutcome{
			Accepted: false,
			Phase:    TxPhasePrepare,
			Reason:   fmt.Sprintf("error: %v", err),
			Err:      err,
		}
	}
	if len(prepareRes.Txs) == 0 {
		return TxLifecycleOutcome{
			Accepted: false,
			Phase:    TxPhasePrepare,
			Reason:   "returned no txs",
		}
	}

	processRes, err := app.ProcessProposal(&abci.RequestProcessProposal{
		Height: ctx.BlockHeight(),
		Time:   ctx.BlockTime(),
		Txs:    prepareRes.Txs,
	})
	if err != nil {
		return TxLifecycleOutcome{
			Accepted: false,
			Phase:    TxPhaseProcess,
			Reason:   fmt.Sprintf("error: %v", err),
			Err:      err,
		}
	}
	if processRes.Status != abci.ResponseProcessProposal_ACCEPT {
		return TxLifecycleOutcome{
			Accepted: false,
			Phase:    TxPhaseProcess,
			Reason:   fmt.Sprintf("status=%s", processRes.Status.String()),
		}
	}

	_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
	if err != nil {
		return TxLifecycleOutcome{
			Accepted: false,
			Phase:    TxPhaseFinalize,
			Reason:   fmt.Sprintf("error: %v", err),
			Err:      err,
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
