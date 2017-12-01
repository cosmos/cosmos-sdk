package search

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/cosmos/cosmos-sdk/client/commands"

	"github.com/tendermint/go-wire/data"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// FindTx performs the given search
func FindTx(query string, prove bool) ([]*ctypes.ResultTx, error) {
	client := commands.GetNode()
	// TODO: actually verify these proofs!!!
	return client.TxSearch(query, prove)
}

// FindAnyTx search all of the strings sequentionally and
// returns combined result
func FindAnyTx(prove bool, queries ...string) ([]*ctypes.ResultTx, error) {
	var all []*ctypes.ResultTx
	// combine all requests
	for _, q := range queries {
		txs, err := FindTx(q, prove)
		if err != nil {
			return nil, err
		}
		all = append(all, txs...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Height > all[j].Height
	})
	return all, nil
}

// TxExtractor take an encoded tx and converts it to
// a usable format
type TxExtractor func([]byte) (interface{}, error)

// TxOutput is like the query output, but stores Tx data
type TxOutput struct {
	Tx     interface{} `json:"tx"`
	Height int64       `json:"height"`
}

// FormatSearch takes the FindTx results and converts them
// to a format to display with help of a TxExtractor that
// knows how to present the tx
func FormatSearch(res []*ctypes.ResultTx, fn TxExtractor) ([]TxOutput, error) {
	out := make([]TxOutput, 0, len(res))
	for _, r := range res {
		ctx, err := fn(r.Tx)
		if err != nil {
			return nil, err
		}
		ro := TxOutput{
			Height: int64(r.Height),
			Tx:     ctx,
		}
		out = append(out, ro)
	}
	return out, nil
}

// Foutput writes the output of wrapping height and info
// in the form {"data": <the_data>, "height": <the_height>}
// to the provider io.Writer
func Foutput(w io.Writer, v interface{}) error {
	blob, err := data.ToJSON(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", string(blob))
	return err
}

// Output prints the search results from FormatSearch
// to stdout
func Output(data interface{}) error {
	return Foutput(os.Stdout, data)
}
