package rest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/gorilla/mux"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, kb keys.Keybase) {
	r.HandleFunc(
		"/stake/delegators/{delegatorAddr}/delegations",
		delegationsRequestHandlerFn(cdc, kb, cliCtx),
	).Methods("POST")
}

type msgDelegationsInput struct {
	DelegatorAddr string   `json:"delegator_addr"` // in bech32
	ValidatorAddr string   `json:"validator_addr"` // in bech32
	Delegation    sdk.Coin `json:"delegation"`
}
type msgBeginRedelegateInput struct {
	DelegatorAddr    string `json:"delegator_addr"`     // in bech32
	ValidatorSrcAddr string `json:"validator_src_addr"` // in bech32
	ValidatorDstAddr string `json:"validator_dst_addr"` // in bech32
	SharesAmount     string `json:"shares"`
}
type msgCompleteRedelegateInput struct {
	DelegatorAddr    string `json:"delegator_addr"`     // in bech32
	ValidatorSrcAddr string `json:"validator_src_addr"` // in bech32
	ValidatorDstAddr string `json:"validator_dst_addr"` // in bech32
}
type msgBeginUnbondingInput struct {
	DelegatorAddr string `json:"delegator_addr"` // in bech32
	ValidatorAddr string `json:"validator_addr"` // in bech32
	SharesAmount  string `json:"shares"`
}
type msgCompleteUnbondingInput struct {
	DelegatorAddr string `json:"delegator_addr"` // in bech32
	ValidatorAddr string `json:"validator_addr"` // in bech32
}

// the request body for edit delegations
type EditDelegationsBody struct {
	LocalAccountName    string                       `json:"name"`
	Password            string                       `json:"password"`
	ChainID             string                       `json:"chain_id"`
	AccountNumber       int64                        `json:"account_number"`
	Sequence            int64                        `json:"sequence"`
	Gas                 string                       `json:"gas"`
	GasAdjustment       string                       `json:"gas_adjustment"`
	Delegations         []msgDelegationsInput        `json:"delegations"`
	BeginUnbondings     []msgBeginUnbondingInput     `json:"begin_unbondings"`
	CompleteUnbondings  []msgCompleteUnbondingInput  `json:"complete_unbondings"`
	BeginRedelegates    []msgBeginRedelegateInput    `json:"begin_redelegates"`
	CompleteRedelegates []msgCompleteRedelegateInput `json:"complete_redelegates"`
}

// nolint: gocyclo
// TODO: Split this up into several smaller functions, and remove the above nolint
// TODO: use sdk.ValAddress instead of sdk.AccAddress for validators in messages
func delegationsRequestHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m EditDelegationsBody

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		err = cdc.UnmarshalJSON(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// build messages
		messages := make([]sdk.Msg, len(m.Delegations)+
			len(m.BeginRedelegates)+
			len(m.CompleteRedelegates)+
			len(m.BeginUnbondings)+
			len(m.CompleteUnbondings))

		i := 0
		for _, msg := range m.Delegations {
			delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode delegator. Error: %s", err.Error()))
				return
			}

			valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
				return
			}

			if !bytes.Equal(info.GetPubKey().Address(), delAddr) {
				utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
				return
			}

			messages[i] = stake.MsgDelegate{
				DelegatorAddr: delAddr,
				ValidatorAddr: valAddr,
				Delegation:    msg.Delegation,
			}

			i++
		}

		for _, msg := range m.BeginRedelegates {
			delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
				return
			}

			if !bytes.Equal(info.GetPubKey().Address(), delAddr) {
				utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
				return
			}

			valSrcAddr, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
				return
			}
			valDstAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
				return
			}

			shares, err := sdk.NewDecFromStr(msg.SharesAmount)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode shares amount. Error: %s", err.Error()))
				return
			}

			messages[i] = stake.MsgBeginRedelegate{
				DelegatorAddr:    delAddr,
				ValidatorSrcAddr: valSrcAddr,
				ValidatorDstAddr: valDstAddr,
				SharesAmount:     shares,
			}

			i++
		}

		for _, msg := range m.CompleteRedelegates {
			delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode delegator. Error: %s", err.Error()))
				return
			}

			valSrcAddr, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
				return
			}

			valDstAddr, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
				return
			}

			if !bytes.Equal(info.GetPubKey().Address(), delAddr) {
				utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
				return
			}

			messages[i] = stake.MsgCompleteRedelegate{
				DelegatorAddr:    delAddr,
				ValidatorSrcAddr: valSrcAddr,
				ValidatorDstAddr: valDstAddr,
			}

			i++
		}

		for _, msg := range m.BeginUnbondings {
			delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode delegator. Error: %s", err.Error()))
				return
			}

			if !bytes.Equal(info.GetPubKey().Address(), delAddr) {
				utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
				return
			}

			valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
				return
			}

			shares, err := sdk.NewDecFromStr(msg.SharesAmount)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode shares amount. Error: %s", err.Error()))
				return
			}

			messages[i] = stake.MsgBeginUnbonding{
				DelegatorAddr: delAddr,
				ValidatorAddr: valAddr,
				SharesAmount:  shares,
			}

			i++
		}

		for _, msg := range m.CompleteUnbondings {
			delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode delegator. Error: %s", err.Error()))
				return
			}

			valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddr)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
				return
			}

			if !bytes.Equal(info.GetPubKey().Address(), delAddr) {
				utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own delegator address")
				return
			}

			messages[i] = stake.MsgCompleteUnbonding{
				DelegatorAddr: delAddr,
				ValidatorAddr: valAddr,
			}

			i++
		}

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
			ChainID:       m.ChainID,
		}

		// sign messages
		signedTxs := make([][]byte, len(messages[:]))
		for i, msg := range messages {
			// increment sequence for each message
			txBldr = txBldr.WithAccountNumber(m.AccountNumber)
			txBldr = txBldr.WithSequence(m.Sequence)
			m.Sequence++

			if utils.HasDryRunArg(r) || txBldr.SimulateGas {
				newBldr, err := utils.EnrichCtxWithGas(txBldr, cliCtx, m.LocalAccountName, []sdk.Msg{msg})
				if err != nil {
					utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
					return
				}
				if utils.HasDryRunArg(r) {
					utils.WriteSimulationResponse(w, newBldr.Gas)
					return
				}
				txBldr = newBldr
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

			signedTxs[i] = txBytes
		}

		// send
		// XXX the operation might not be atomic if a tx fails
		//     should we have a sdk.MultiMsg type to make sending atomic?
		results := make([]*ctypes.ResultBroadcastTxCommit, len(signedTxs[:]))
		for i, txBytes := range signedTxs {
			res, err := cliCtx.BroadcastTx(txBytes)
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}

			results[i] = res
		}

		output, err := codec.MarshalJSONIndent(cdc, results[:])
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Write(output)
	}
}
