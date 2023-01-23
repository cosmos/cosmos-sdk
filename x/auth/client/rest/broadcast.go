package rest

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

// BroadcastReq defines a tx broadcasting request.
type BroadcastReq struct {
	Tx         legacytx.StdTx `json:"tx" yaml:"tx"`
	Mode       string         `json:"mode" yaml:"mode"`
	Sequences  []uint64       `json:"sequences" yaml:"sequences"`
	FeeGranter string         `json:"fee_granter" yaml:"fee_granter"`
}

var _ codectypes.UnpackInterfacesMessage = BroadcastReq{}

// UnpackInterfaces implements the UnpackInterfacesMessage interface.
func (m BroadcastReq) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Tx.UnpackInterfaces(unpacker)
}

// BroadcastTxRequest implements a tx broadcasting handler that is responsible
// for broadcasting a valid and signed tx to a full node. The tx can be
// broadcasted via a sync|async|block mechanism.
func BroadcastTxRequest(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BroadcastReq

		body, err := io.ReadAll(r.Body)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		// NOTE: amino is used intentionally here, don't migrate it!
		err = clientCtx.LegacyAmino.UnmarshalJSON(body, &req)
		if err != nil {
			err := fmt.Errorf("this transaction cannot be broadcasted via legacy REST endpoints, because it does not support"+
				" Amino serialization. Please either use CLI, gRPC, gRPC-gateway, or directly query the Tendermint RPC"+
				" endpoint to broadcast this transaction. The new REST endpoint (via gRPC-gateway) is POST /cosmos/tx/v1beta1/txs."+
				" Please also see the REST endpoints migration guide at %s for more info", clientrest.DeprecationURL)
			if rest.CheckBadRequestError(w, err) {
				return
			}
		}

		txBuilder := clientCtx.TxConfig.NewTxBuilder()
		txBuilder.SetFeeAmount(req.Tx.GetFee())
		txBuilder.SetGasLimit(req.Tx.GetGas())
		txBuilder.SetMemo(req.Tx.GetMemo())
		if err := txBuilder.SetMsgs(req.Tx.GetMsgs()...); rest.CheckBadRequestError(w, err) {
			return
		}

		txBuilder.SetTimeoutHeight(req.Tx.GetTimeoutHeight())
		if req.FeeGranter != "" {
			addr, err := sdk.AccAddressFromBech32(req.FeeGranter)
			if rest.CheckBadRequestError(w, err) {
				return
			}

			txBuilder.SetFeeGranter(addr)
		}

		signatures, err := req.Tx.GetSignaturesV2()
		if rest.CheckBadRequestError(w, err) {
			return
		}

		// if sequence is not given, try fetch from the chain
		if len(req.Sequences) == 0 {
			for _, sig := range signatures {
				_, seq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, sdk.AccAddress(sig.PubKey.Address().Bytes()))
				if rest.CheckBadRequestError(w, err) {
					return
				}
				req.Sequences = append(req.Sequences, seq)
			}
		}

		// check the sequence nubmer is equal with the signature nubmer
		if len(signatures) != len(req.Sequences) {
			rest.CheckBadRequestError(w, errors.New("must provide each signers's sequence number"))
			return
		}

		// fill sequence number to new signature
		for i, seq := range req.Sequences {
			signatures[i].Sequence = seq
		}

		if err := txBuilder.SetSignatures(signatures...); rest.CheckBadRequestError(w, err) {
			return
		}

		// compute signature bytes
		txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithBroadcastMode(req.Mode)
		res, err := clientCtx.BroadcastTx(txBytes)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		rest.PostProcessResponseBare(w, clientCtx, res)
	}
}
