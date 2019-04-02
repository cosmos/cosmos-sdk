package tx

import (
	"encoding/base64"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

type (
	// DecodeReq defines a tx decoding request.
	DecodeReq struct {
		Tx string `json:"tx"`
	}

	// DecodeResp defines a tx decoding response.
	DecodeResp struct {
		Tx auth.StdTx `json:"tx"`
	}
)

// DecodeTxRequestHandlerFn returns the decode tx REST handler. In particular,
// it takes base64-decoded bytes, decodes it from the Amino wire protocol,
// and responds with a json-formatted transaction.
func DecodeTxRequestHandlerFn(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DecodeReq

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		err = cdc.UnmarshalJSON(body, &req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		txBytes, err := base64.StdEncoding.DecodeString(req.Tx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var stdTx auth.StdTx
		err = cliCtx.Codec.UnmarshalBinaryLengthPrefixed(txBytes, &stdTx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		response := DecodeResp{Tx: stdTx}
		rest.PostProcessResponse(w, cdc, response, cliCtx.Indent)
	}
}

// txDecodeRespStr implements a simple Stringer wrapper for a decoded tx.
type txDecodeRespTx auth.StdTx

func (tx txDecodeRespTx) String() string {
	return tx.String()
}

// GetDecodeCommand returns the decode command to take Amino-serialized bytes and turn it into
// a JSONified transaction
func GetDecodeCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode [amino-byte-string]",
		Short: "Decode an amino-encoded transaction string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliCtx := context.NewCLIContext().WithCodec(codec)

			txBytesBase64 := args[0]

			txBytes, err := base64.StdEncoding.DecodeString(txBytesBase64)
			if err != nil {
				return err
			}

			var stdTx auth.StdTx
			err = cliCtx.Codec.UnmarshalBinaryLengthPrefixed(txBytes, &stdTx)
			if err != nil {
				return err
			}

			response := txDecodeRespTx(stdTx)
			_ = cliCtx.PrintOutput(response)

			return nil
		},
	}

	return client.PostCommands(cmd)[0]
}
