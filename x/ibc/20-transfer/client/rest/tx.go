package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	clientutils "github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/utils"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	channelutils "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/transfer", RestPortID, RestChannelID), transferHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/packet/receive"), recvPacketHandlerFn(cliCtx)).Methods("POST")
}

func transferHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]

		var req TransferTxReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgTransfer(
			portID,
			channelID,
			req.Amount,
			fromAddr,
			req.Receiver,
			req.Source,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

/*
func recvPacketHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RecvPacketReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgRecvPacket(
			req.Packet,
			req.Proofs,
			req.Height,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
*/

func recvPacketHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cliCtx.Codec))
		cliCtx := context.NewCLIContext().WithCodec(cliCtx.Codec).WithBroadcastMode(flags.BroadcastBlock)

		var req RecvPacketReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		sourceNode := req.SourceNode
		sourceChainID := req.SourceChainID

		cliCtx2 := context.NewCLIContextIBC(req.BaseReq.From, sourceChainID, sourceNode).
			WithCodec(cliCtx.Codec).
			WithBroadcastMode(flags.BroadcastBlock)

		header, err := clientutils.GetTendermintHeader(cliCtx2)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		clientid := req.ClientID
		sourcePort := req.SourcePortID
		sourceChannel := req.SourceChannelID

		passphrase, err := keys.GetPassphrase(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msgUpdateClient := clienttypes.NewMsgUpdateClient(clientid, header, fromAddr)
		if err := msgUpdateClient.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := utils.CompleteAndBroadcastTx(txBldr, cliCtx, []sdk.Msg{msgUpdateClient}, passphrase)
		if err != nil || !res.IsOK() {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		sequence, err := strconv.ParseUint(req.Sequence, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		timeout, err := strconv.ParseUint(req.Timeout, 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		packetRes, err := channelutils.QueryPacket(cliCtx2.WithHeight(header.Height-1), sourcePort, sourceChannel, sequence, timeout, "ibc")
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgRecvPacket(
			packetRes.Packet,
			[]commitment.Proof{packetRes.Proof},
			packetRes.ProofHeight,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
