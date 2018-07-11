package tx

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	flagTags = "tag"
	flagAny  = "any"
)

// default client command to search through tagged transactions
func SearchTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Search for all transactions that match the given tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags := viper.GetStringSlice(flagTags)

			txs, err := searchTxs(context.NewCoreContextFromViper(), cdc, tags)
			if err != nil {
				return err
			}
			output, err := cdc.MarshalJSON(txs)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")

	// TODO: change this to false once proofs built in
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	cmd.Flags().StringSlice(flagTags, nil, "Tags that must match (may provide multiple)")
	cmd.Flags().Bool(flagAny, false, "Return transactions that match ANY tag, rather than ALL")
	return cmd
}

func searchTxs(ctx context.CoreContext, cdc *wire.Codec, tags []string) ([]txInfo, error) {
	if len(tags) == 0 {
		return nil, errors.New("must declare at least one tag to search")
	}
	// XXX: implement ANY
	query := strings.Join(tags, " AND ")
	// get the node
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	prove := !viper.GetBool(client.FlagTrustNode)
	// TODO: take these as args
	page := 0
	perPage := 100
	res, err := node.TxSearch(query, prove, page, perPage)
	if err != nil {
		return nil, err
	}

	info, err := formatTxResults(cdc, res.Txs)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func formatTxResults(cdc *wire.Codec, res []*ctypes.ResultTx) ([]txInfo, error) {
	var err error
	out := make([]txInfo, len(res))
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
func SearchTxRequestHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tag := r.FormValue("tag")
		if tag == "" {
			w.WriteHeader(400)
			w.Write([]byte("You need to provide at least a tag as a key=value pair to search for. Postfix the key with _bech32 to search bech32-encoded addresses or public keys"))
			return
		}
		keyValue := strings.Split(tag, "=")
		key := keyValue[0]
		value, err := url.QueryUnescape(keyValue[1])
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Could not decode address: " + err.Error()))
			return
		}
		if strings.HasSuffix(key, "_bech32") {
			bech32address := strings.Trim(value, "'")
			prefix := strings.Split(bech32address, "1")[0]
			bz, err := sdk.GetFromBech32(bech32address, prefix)
			if err != nil {
				w.WriteHeader(400)
				w.Write([]byte(err.Error()))
				return
			}

			tag = strings.TrimRight(key, "_bech32") + "='" + sdk.AccAddress(bz).String() + "'"
		}

		txs, err := searchTxs(ctx, cdc, []string{tag})
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		if len(txs) == 0 {
			w.Write([]byte("[]"))
			return
		}

		output, err := cdc.MarshalJSON(txs)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
