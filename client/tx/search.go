package tx

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	flagTags = "tags"
	flagAny  = "any"
)

// default client command to search through tagged transactions
func SearchTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Search for all transactions that match the given tags.",
		Long: strings.TrimSpace(`
Search for transactions that match exactly the given tags. For example:

$ gaiacli query txs --tags '<tag1>:<value1>&<tag2>:<value2>'
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			tagsStr := viper.GetString(flagTags)
			tagsStr = strings.Trim(tagsStr, "'")
			var tags []string
			if strings.Contains(tagsStr, "&") {
				tags = strings.Split(tagsStr, "&")
			} else {
				tags = append(tags, tagsStr)
			}

			var tmTags []string
			for _, tag := range tags {
				if !strings.Contains(tag, ":") {
					return fmt.Errorf("%s should be of the format <key>:<value>", tagsStr)
				} else if strings.Count(tag, ":") > 1 {
					return fmt.Errorf("%s should only contain one <key>:<value> pair", tagsStr)
				}

				keyValue := strings.Split(tag, ":")
				tag = fmt.Sprintf("%s='%s'", keyValue[0], keyValue[1])
				tmTags = append(tmTags, tag)
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txs, err := searchTxs(cliCtx, cdc, tmTags)
			if err != nil {
				return err
			}

			var output []byte
			if cliCtx.Indent {
				output, err = cdc.MarshalJSONIndent(txs, "", "  ")
			} else {
				output, err = cdc.MarshalJSON(txs)
			}

			if err != nil {
				return err
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(client.FlagNode, cmd.Flags().Lookup(client.FlagNode))
	cmd.Flags().String(client.FlagChainID, "", "Chain ID of Tendermint node")
	viper.BindPFlag(client.FlagChainID, cmd.Flags().Lookup(client.FlagChainID))
	cmd.Flags().Bool(client.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	viper.BindPFlag(client.FlagTrustNode, cmd.Flags().Lookup(client.FlagTrustNode))
	cmd.Flags().String(flagTags, "", "tag:value list of tags that must match")
	return cmd
}

func searchTxs(cliCtx context.CLIContext, cdc *codec.Codec, tags []string) ([]Info, error) {
	if len(tags) == 0 {
		return nil, errors.New("must declare at least one tag to search")
	}

	// XXX: implement ANY
	query := strings.Join(tags, " AND ")

	// get the node
	node, err := cliCtx.GetNode()
	if err != nil {
		return nil, err
	}

	prove := !cliCtx.TrustNode

	// TODO: take these as args
	page := 0
	perPage := 100
	res, err := node.TxSearch(query, prove, page, perPage)
	if err != nil {
		return nil, err
	}

	if prove {
		for _, tx := range res.Txs {
			err := ValidateTxResult(cliCtx, tx)
			if err != nil {
				return nil, err
			}
		}
	}

	info, err := FormatTxResults(cdc, res.Txs)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// parse the indexed txs into an array of Info
func FormatTxResults(cdc *codec.Codec, res []*ctypes.ResultTx) ([]Info, error) {
	var err error
	out := make([]Info, len(res))
	for i := range res {
		out[i], err = formatTxResult(cdc, res[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

/////////////////////////////////////////
// REST

// Search Tx REST Handler
func SearchTxRequestHandlerFn(cliCtx context.CLIContext, cdc *codec.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tags []string
		var txs []Info
		err := r.ParseForm()
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}
		if len(r.Form) == 0 {
			utils.PostProcessResponse(w, cdc, txs, cliCtx.Indent)
			return
		}

		for key, values := range r.Form {
			value, err := url.QueryUnescape(values[0])
			if err != nil {
				utils.WriteErrorResponse(w, http.StatusBadRequest, sdk.AppendMsgToErr("could not decode query value", err.Error()))
				return
			}

			tag := fmt.Sprintf("%s='%s'", key, value)
			tags = append(tags, tag)
		}

		txs, err = searchTxs(cliCtx, cdc, tags)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.PostProcessResponse(w, cdc, txs, cliCtx.Indent)
	}
}
