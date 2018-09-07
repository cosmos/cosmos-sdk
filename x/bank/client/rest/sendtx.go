package rest

import (
	"io/ioutil"
	"net/http"

	cliclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/client"

	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/accounts/{address}/send", SendRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
}

type sendBody struct {
	// fees is not used currently
	// Fees             sdk.Coin  `json="fees"`
	Amount           sdk.Coins `json:"amount"`
	LocalAccountName string    `json:"name"`
	Password         string    `json:"password"`
	ChainID          string    `json:"chain_id"`
	AccountNumber    int64     `json:"account_number"`
	Sequence         int64     `json:"sequence"`
	Gas              int64     `json:"gas"`
	GasAdjustment    string    `json:"gas_adjustment"`
}

var msgCdc = wire.NewCodec()

func init() {
	bank.RegisterWire(msgCdc)
}

// SendRequestHandlerFn - http request handler to send coins to a address
// nolint: gocyclo
func SendRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		bech32addr := vars["address"]

		to, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var m sendBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		err = msgCdc.UnmarshalJSON(body, &m)
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
		msg := client.BuildMsg(sdk.AccAddress(info.GetPubKey().Address()), to, m.Amount)
		if err != nil { // XXX rechecking same error ?
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		txCtx := authctx.TxContext{
			Codec:         cdc,
			Gas:           m.Gas,
			ChainID:       m.ChainID,
			AccountNumber: m.AccountNumber,
			Sequence:      m.Sequence,
		}

		adjustment, ok := utils.ParseFloat64OrReturnBadRequest(w, m.GasAdjustment, cliclient.DefaultGasAdjustment)
		if !ok {
			return
		}
		cliCtx = cliCtx.WithGasAdjustment(adjustment)

		if utils.HasDryRunArg(r) || m.Gas == 0 {
			newCtx, err := utils.EnrichCtxWithGas(txCtx, cliCtx, m.LocalAccountName, []sdk.Msg{msg})
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			if utils.HasDryRunArg(r) {
				utils.WriteSimulationResponse(w, txCtx.Gas)
				return
			}
			txCtx = newCtx
		}

		if utils.HasGenerateOnlyArg(r) {
			utils.WriteGenerateStdTxResponse(w, txCtx, []sdk.Msg{msg})
			return
		}

		txBytes, err := txCtx.BuildAndSign(m.LocalAccountName, m.Password, []sdk.Msg{msg})
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		res, err := cliCtx.BroadcastTx(txBytes)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		output, err := wire.MarshalJSONIndent(cdc, res)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Write(output)
	}
}
