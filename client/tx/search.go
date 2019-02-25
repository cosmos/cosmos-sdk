package tx

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

const (
	flagTags     = "tags"
	flagAny      = "any"
	flagPage     = "page"
	flagLimit    = "limit"
	defaultPage  = 1
	defaultLimit = 30 // should be consistent with tendermint/tendermint/rpc/core/pipe.go:19
)

// default client command to search through tagged transactions
func SearchTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Search for all transactions that match the given tags.",
		Long: strings.TrimSpace(`
Search for transactions that match exactly the given tags. For example:

$ gaiacli query txs --tags '<tag1>:<value1>&<tag2>:<value2>' --page 1 --limit 30
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
				if keyValue[0] == types.TxHeightKey {
					tag = fmt.Sprintf("%s=%s", keyValue[0], keyValue[1])
				} else {
					tag = fmt.Sprintf("%s='%s'", keyValue[0], keyValue[1])
				}
				tmTags = append(tmTags, tag)
			}
			page := viper.GetInt(flagPage)
			limit := viper.GetInt(flagLimit)

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txs, err := SearchTxs(cliCtx, cdc, tmTags, page, limit)
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
	cmd.Flags().Bool(client.FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	viper.BindPFlag(client.FlagTrustNode, cmd.Flags().Lookup(client.FlagTrustNode))
	cmd.Flags().String(flagTags, "", "tag:value list of tags that must match")
	cmd.Flags().Int32(flagPage, defaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int32(flagLimit, defaultLimit, "Query number of transactions results per page returned")
	cmd.MarkFlagRequired(flagTags)
	return cmd
}

// SearchTxs performs a search for transactions for a given set of tags via
// Tendermint RPC. It returns a slice of Info object containing txs and metadata.
// An error is returned if the query fails.
func SearchTxs(cliCtx context.CLIContext, cdc *codec.Codec, tags []string, page, limit int) ([]sdk.TxResponse, error) {
	if len(tags) == 0 {
		return nil, errors.New("must declare at least one tag to search")
	}

	if page <= 0 {
		return nil, errors.New("page must greater than 0")
	}

	if limit <= 0 {
		return nil, errors.New("limit must greater than 0")
	}

	// XXX: implement ANY
	query := strings.Join(tags, " AND ")

	// get the node
	node, err := cliCtx.GetNode()
	if err != nil {
		return nil, err
	}

	prove := !cliCtx.TrustNode

	res, err := node.TxSearch(query, prove, page, limit)
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
func FormatTxResults(cdc *codec.Codec, res []*ctypes.ResultTx) ([]sdk.TxResponse, error) {
	var err error
	out := make([]sdk.TxResponse, len(res))
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
		var page, limit int
		var txs []sdk.TxResponse
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}
		if len(r.Form) == 0 {
			rest.PostProcessResponse(w, cdc, txs, cliCtx.Indent)
			return
		}

		tags, page, limit, err = parseHTTPArgs(r)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		txs, err = SearchTxs(cliCtx, cdc, tags, page, limit)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, txs, cliCtx.Indent)
	}
}

func parseHTTPArgs(r *http.Request) (tags []string, page, limit int, err error) {
	tags = make([]string, 0, len(r.Form))
	for key, values := range r.Form {
		if key == "page" || key == "limit" {
			continue
		}
		var value string
		value, err = url.QueryUnescape(values[0])
		if err != nil {
			return tags, page, limit, err
		}

		var tag string
		if key == types.TxHeightKey {
			tag = fmt.Sprintf("%s=%s", key, value)
		} else {
			tag = fmt.Sprintf("%s='%s'", key, value)
		}
		tags = append(tags, tag)
	}

	pageStr := r.FormValue("page")
	if pageStr == "" {
		page = defaultPage
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return tags, page, limit, err
		} else if page <= 0 {
			return tags, page, limit, errors.New("page must greater than 0")
		}
	}

	limitStr := r.FormValue("limit")
	if limitStr == "" {
		limit = defaultLimit
	} else {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return tags, page, limit, err
		} else if limit <= 0 {
			return tags, page, limit, errors.New("limit must greater than 0")
		}
	}

	return tags, page, limit, nil
}
