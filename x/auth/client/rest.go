package client

import (
	"log"
	"net/http"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/client/flags"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/types/errors"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	"github.com/KiraCore/cosmos-sdk/x/auth/types"
)

// WriteGenerateStdTxResponse writes response for the generate only mode.
func WriteGenerateStdTxResponse(w http.ResponseWriter, clientCtx client.Context, br rest.BaseReq, msgs []sdk.Msg) {
	gasAdj, ok := rest.ParseFloat64OrReturnBadRequest(w, br.GasAdjustment, flags.DefaultGasAdjustment)
	if !ok {
		return
	}

	gasSetting, err := flags.ParseGasSetting(br.Gas)
	if rest.CheckBadRequestError(w, err) {
		return
	}

	txBldr := types.NewTxBuilder(
		GetTxEncoder(clientCtx.Codec), br.AccountNumber, br.Sequence, gasSetting.Gas, gasAdj,
		br.Simulate, br.ChainID, br.Memo, br.Fees, br.GasPrices,
	)

	if br.Simulate || gasSetting.Simulate {
		if gasAdj < 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.ErrorInvalidGasAdjustment.Error())
			return
		}

		txBldr, err = EnrichWithGas(txBldr, clientCtx, msgs)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		if br.Simulate {
			rest.WriteSimulationResponse(w, clientCtx.Codec, txBldr.Gas())
			return
		}
	}

	stdMsg, err := txBldr.BuildSignMsg(msgs)
	if rest.CheckBadRequestError(w, err) {
		return
	}

	output, err := clientCtx.JSONMarshaler.MarshalJSON(types.NewStdTx(stdMsg.Msgs, stdMsg.Fee, nil, stdMsg.Memo))
	if rest.CheckInternalServerError(w, err) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(output); err != nil {
		log.Printf("could not write response: %v", err)
	}

}
