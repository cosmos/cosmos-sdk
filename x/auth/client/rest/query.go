package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	genutilrest "github.com/cosmos/cosmos-sdk/x/genutil/client/rest"
)

// QueryAccountRequestHandlerFn is the query accountREST Handler.
func QueryAccountRequestHandlerFn(storeName string, clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32addr := vars["address"]

		addr, err := sdk.AccAddressFromBech32(bech32addr)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		accGetter := types.AccountRetriever{}

		account, height, err := accGetter.GetAccountWithHeight(clientCtx, addr)
		if err != nil {
			// TODO: Handle more appropriately based on the error type.
			// Ref: https://github.com/cosmos/cosmos-sdk/issues/4923
			if err := accGetter.EnsureExists(clientCtx, addr); err != nil {
				clientCtx = clientCtx.WithHeight(height)
				rest.PostProcessResponse(w, clientCtx, types.BaseAccount{})
				return
			}

			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, account)
	}
}

// QueryTxsRequestHandlerFn implements a REST handler that searches for transactions.
// Genesis transactions are returned if the height parameter is set to zero,
// otherwise the transactions are searched for by events.
func QueryTxsRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(
				w, http.StatusBadRequest,
				fmt.Sprintf("failed to parse query parameters: %s", err),
			)
			return
		}

		// if the height query param is set to zero, query for genesis transactions
		heightStr := r.FormValue("height")
		if heightStr != "" {
			if height, err := strconv.ParseInt(heightStr, 10, 64); err == nil && height == 0 {
				genutilrest.QueryGenesisTxs(clientCtx, w)
				return
			}
		}

		var (
			events      []string
			txs         []sdk.TxResponse
			page, limit int
		)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		if len(r.Form) == 0 {
			rest.PostProcessResponseBare(w, clientCtx, txs)
			return
		}

		events, page, limit, err = rest.ParseHTTPArgs(r)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		searchResult, err := authtx.QueryTxsByEvents(clientCtx, events, page, limit, "")
		if rest.CheckInternalServerError(w, err) {
			return
		}

		for _, txRes := range searchResult.Txs {
			packStdTxResponse(w, clientCtx, txRes)
		}

		err = checkAminoMarshalError(clientCtx, searchResult, "/cosmos/tx/v1beta1/txs")
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())

			return
		}

		rest.PostProcessResponseBare(w, clientCtx, searchResult)
	}
}

// QueryTxRequestHandlerFn implements a REST handler that queries a transaction
// by hash in a committed block.
func QueryTxRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hashHexStr := vars["hash"]

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		output, err := authtx.QueryTx(clientCtx, hashHexStr)
		if err != nil {
			if strings.Contains(err.Error(), hashHexStr) {
				rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
				return
			}
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		err = packStdTxResponse(w, clientCtx, output)
		if err != nil {
			// Error is already returned by packStdTxResponse.
			return
		}

		if output.Empty() {
			rest.WriteErrorResponse(w, http.StatusNotFound, fmt.Sprintf("no transaction found with hash %s", hashHexStr))
		}

		err = checkAminoMarshalError(clientCtx, output, "/cosmos/tx/v1beta1/txs/{txhash}")
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())

			return
		}

		rest.PostProcessResponseBare(w, clientCtx, output)
	}
}

func queryParamsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)
		res, height, err := clientCtx.QueryWithData(route, nil)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

// packStdTxResponse takes a sdk.TxResponse, converts the Tx into a StdTx, and
// packs the StdTx again into the sdk.TxResponse Any. Amino then takes care of
// seamlessly JSON-outputting the Any.
func packStdTxResponse(w http.ResponseWriter, clientCtx client.Context, txRes *sdk.TxResponse) error {
	// We just unmarshalled from Tendermint, we take the proto Tx's raw
	// bytes, and convert them into a StdTx to be displayed.
	txBytes := txRes.Tx.Value
	stdTx, err := convertToStdTx(w, clientCtx, txBytes)
	if err != nil {
		return err
	}

	// Pack the amino stdTx into the TxResponse's Any.
	txRes.Tx = codectypes.UnsafePackAny(stdTx)

	return nil
}

// checkAminoMarshalError checks if there are errors with marshalling non-amino
// txs with amino.
func checkAminoMarshalError(ctx client.Context, resp interface{}, grpcEndPoint string) error {
	// LegacyAmino used intentionally here to handle the SignMode errors
	marshaler := ctx.LegacyAmino

	_, err := marshaler.MarshalJSON(resp)
	if err != nil {

		// If there's an unmarshalling error, we assume that it's because we're
		// using amino to unmarshal a non-amino tx.
		return fmt.Errorf("this transaction cannot be displayed via legacy REST endpoints, because it does not support"+
			" Amino serialization. Please either use CLI, gRPC, gRPC-gateway, or directly query the Tendermint RPC"+
			" endpoint to query this transaction. The new REST endpoint (via gRPC-gateway) is %s. Please also see the"+
			"REST endpoints migration guide at %s for more info", grpcEndPoint, clientrest.DeprecationURL)

	}

	return nil
}
