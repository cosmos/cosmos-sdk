package utils

import (
	"log"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// WriteGenerateStdTxResponse writes response for the generate only mode.
func WriteGenerateStdTxResponse(w http.ResponseWriter, cliCtx context.CLIContext, br rest.BaseReq, msgs []sdk.Msg) {
	gasAdj, ok := rest.ParseFloat64OrReturnBadRequest(w, br.GasAdjustment, flags.DefaultGasAdjustment)
	if !ok {
		return
	}

	simAndExec, gas, err := flags.ParseGas(br.Gas)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	txBldr := types.NewTxBuilder(
		GetTxEncoder(cliCtx.Codec), br.AccountNumber, br.Sequence, gas, gasAdj,
		br.Simulate, br.ChainID, br.Memo, br.Fees, br.GasPrices,
	)

	if br.Simulate || simAndExec {
		if gasAdj < 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errInvalidGasAdjustment.Error())
			return
		}

		txBldr, err = EnrichWithGas(txBldr, cliCtx, msgs)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if br.Simulate {
			rest.WriteSimulationResponse(w, cliCtx.Codec, txBldr.Gas())
			return
		}
	}

	stdMsg, err := txBldr.BuildSignMsg(msgs)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := cliCtx.Codec.MarshalJSON(types.NewStdTx(stdMsg.Msgs, stdMsg.Fee, nil, stdMsg.Memo))
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(output); err != nil {
		log.Printf("could not write response: %v", err)
	}

	return
}
