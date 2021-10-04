package rest

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

// EncodeReq defines a tx encode request.
type EncodeReq struct {
	Tx         legacytx.StdTx `json:"tx" yaml:"tx"`
	Sequences  []uint64       `json:"sequences" yaml:"sequences"`
	FeeGranter string         `json:"fee_granter" yaml:"fee_granter"`
}

// EncodeResq defines a tx encode response.
type EncodeResq struct {
	Tx []byte `json:"tx" yaml:"tx"`
}

var _ codectypes.UnpackInterfacesMessage = EncodeReq{}

// UnpackInterfaces implements the UnpackInterfacesMessage interface.
func (m EncodeReq) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Tx.UnpackInterfaces(unpacker)
}

// EncodeTxRequest implements a tx encode handler that is responsible
// for encoding a legacy tx into new format tx bytes.
func EncodeTxRequest(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EncodeReq

		body, err := ioutil.ReadAll(r.Body)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		// NOTE: amino is used intentionally here, don't migrate it!
		err = clientCtx.LegacyAmino.UnmarshalJSON(body, &req)
		if err != nil {
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
			rest.CheckBadRequestError(w, errors.New("Must provide each signers's sequence number"))
			return
		}

		// fill sequence number to new signature
		for i, seq := range req.Sequences {
			signatures[i].Sequence = seq
		}

		if err := txBuilder.SetSignatures(signatures...); rest.CheckBadRequestError(w, err) {
			return
		}

		// compute tx bytes
		txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
		if rest.CheckInternalServerError(w, err) {
			return
		}

		rest.PostProcessResponseBare(w, clientCtx, EncodeResq{
			Tx: txBytes,
		})
	}
}
