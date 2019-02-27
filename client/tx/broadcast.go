package tx

import (
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	// Returns with the response from CheckTx.
	flagSync = "sync"
	// Returns right away, with no response
	flagAsync = "async"
	// Only returns error if mempool.BroadcastTx errs (ie. problem with the app) or if we timeout waiting for tx to commit.
	flagBlock = "block"
)

// BroadcastReq defines a tx broadcasting request.
type BroadcastReq struct {
	Tx     auth.StdTx `json:"tx"`
	Return string     `json:"return"`
}

// BroadcastTxRequest implements a tx broadcasting handler that is responsible
// for broadcasting a valid and signed tx to a full node. The tx can be
// broadcasted via a sync|async|block mechanism.
func BroadcastTxRequest(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BroadcastReq

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

		txBytes, err := cdc.MarshalBinaryLengthPrefixed(req.Tx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var res interface{}
		switch req.Return {
		case flagBlock:
			res, err = cliCtx.BroadcastTx(txBytes)

		case flagSync:
			res, err = cliCtx.BroadcastTxSync(txBytes)

		case flagAsync:
			res, err = cliCtx.BroadcastTxAsync(txBytes)

		default:
			rest.WriteErrorResponse(w, http.StatusInternalServerError,
				"unsupported return type. supported types: block, sync, async")
			return
		}

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

// GetBroadcastCommand returns the tx broadcast command.
func GetBroadcastCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broadcast [file_path]",
		Short: "Broadcast transactions generated offline",
		Long: strings.TrimSpace(`Broadcast transactions created with the --generate-only
flag and signed with the sign command. Read a transaction from [file_path] and
broadcast it to a node. If you supply a dash (-) argument in place of an input
filename, the command reads from standard input.

$ gaiacli tx broadcast ./mytxn.json
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliCtx := context.NewCLIContext().WithCodec(codec)
			stdTx, err := utils.ReadStdTxFromFile(cliCtx.Codec, args[0])
			if err != nil {
				return
			}

			txBytes, err := cliCtx.Codec.MarshalBinaryLengthPrefixed(stdTx)
			if err != nil {
				return
			}

			res, err := cliCtx.BroadcastTx(txBytes)
			cliCtx.PrintOutput(res)
			return err
		},
	}

	return client.PostCommands(cmd)[0]
}
