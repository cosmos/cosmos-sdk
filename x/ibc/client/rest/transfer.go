package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/ibc"

	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, kb keys.Keybase) {
	r.HandleFunc("/ibc/{destchain}/{address}/send", TransferRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
}

type transferBody struct {
	// Fees             sdk.Coin  `json="fees"`
	Amount           sdk.Coins `json:"amount"`
	LocalAccountName string    `json:"name"`
	Password         string    `json:"password"`
	SrcChainID       string    `json:"src_chain_id"`
	AccountNumber    int64     `json:"account_number"`
	Sequence         int64     `json:"sequence"`
	Gas              string    `json:"gas"`
	GasAdjustment    string    `json:"gas_adjustment"`
}

// TransferRequestHandler - http request handler to transfer coins to a address
// on a different chain via IBC
func TransferRequestHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		destChainID := vars["destchain"]
		bech32addr := vars["address"]

		to, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var m transferBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		err = cdc.UnmarshalJSON(body, &m)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		// build message
		packet := ibc.NewIBCPacket(sdk.AccAddress(info.GetPubKey().Address()), to, m.Amount, m.SrcChainID, destChainID)
		msg := ibc.IBCTransferMsg{packet}

		simulateGas, gas, err := client.ReadGasFlag(m.Gas)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		adjustment, ok := utils.ParseFloat64OrReturnBadRequest(w, m.GasAdjustment, client.DefaultGasAdjustment)
		if !ok {
			return
		}
		txBldr := authtxb.TxBuilder{
			Codec:         cdc,
			Gas:           gas,
			GasAdjustment: adjustment,
			SimulateGas:   simulateGas,
			ChainID:       m.SrcChainID,
			AccountNumber: m.AccountNumber,
			Sequence:      m.Sequence,
		}

		if utils.HasDryRunArg(r) || txBldr.SimulateGas {
			newCtx, err := utils.EnrichCtxWithGas(txBldr, cliCtx, m.LocalAccountName, []sdk.Msg{msg})
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			if utils.HasDryRunArg(r) {
				utils.WriteSimulationResponse(w, txBldr.Gas)
				return
			}
			txBldr = newCtx
		}

		if utils.HasGenerateOnlyArg(r) {
			utils.WriteGenerateStdTxResponse(w, txBldr, []sdk.Msg{msg})
			return
		}

		txBytes, err := txBldr.BuildAndSign(m.LocalAccountName, m.Password, []sdk.Msg{msg})
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		res, err := cliCtx.BroadcastTx(txBytes)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		output, err := cdc.MarshalJSON(res)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Write(output)
	}
}
