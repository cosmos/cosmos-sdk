package tx

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	"cosmossdk.io/x/tx/decode"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

// QueryTxsByEvents retrieves a list of paginated transactions from CometBFT's
// TxSearch RPC method given a set of pagination criteria and an events query.
// Note, the events query must be valid based on CometBFT's query semantics.
// An error is returned if the query or parsing fails or if the query is empty.
//
// Note, if an empty orderBy is provided, the default behavior is ascending. If
// negative values are provided for page or limit, defaults will be used.
func QueryTxsByEvents(clientCtx client.Context, page, limit int, query, orderBy string) (*sdk.SearchTxsResult, error) {
	if len(query) == 0 {
		return nil, errors.New("query cannot be empty")
	}

	// CometBFT node.TxSearch that is used for querying txs defines pages
	// starting from 1, so we default to 1 if not provided in the request.
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = querytypes.DefaultLimit
	}

	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	resTxs, err := node.TxSearch(context.Background(), query, false, &page, &limit, orderBy)
	if err != nil {
		return nil, fmt.Errorf("failed to search for txs: %w", err)
	}

	resBlocks, err := getBlocksForTxResults(clientCtx, resTxs.Txs)
	if err != nil {
		return nil, err
	}

	txs, err := formatTxResults(clientCtx, resTxs.Txs, resBlocks)
	if err != nil {
		return nil, err
	}

	return sdk.NewSearchTxsResult(uint64(resTxs.TotalCount), uint64(len(txs)), uint64(page), uint64(limit), txs), nil
}

// QueryTx queries for a single transaction by a hash string in hex format. An
// error is returned if the transaction does not exist or cannot be queried.
func QueryTx(clientCtx client.Context, hashHexStr string) (*sdk.TxResponse, error) {
	hash, err := hex.DecodeString(hashHexStr)
	if err != nil {
		return nil, err
	}

	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	resTx, err := node.Tx(context.Background(), hash, true)
	if err != nil {
		return nil, err
	}

	resBlocks, err := getBlocksForTxResults(clientCtx, []*coretypes.ResultTx{resTx})
	if err != nil {
		return nil, err
	}

	out, err := mkTxResult(clientCtx, resTx, resBlocks[resTx.Height])
	if err != nil {
		return out, err
	}

	return out, nil
}

// formatTxResults parses the indexed txs into a slice of TxResponse objects.
func formatTxResults(clientCtx client.Context, resTxs []*coretypes.ResultTx, resBlocks map[int64]*coretypes.ResultBlock) ([]*sdk.TxResponse, error) {
	var err error
	out := make([]*sdk.TxResponse, len(resTxs))
	for i := range resTxs {
		out[i], err = mkTxResult(clientCtx, resTxs[i], resBlocks[resTxs[i].Height])
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func getBlocksForTxResults(clientCtx client.Context, resTxs []*coretypes.ResultTx) (map[int64]*coretypes.ResultBlock, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	resBlocks := make(map[int64]*coretypes.ResultBlock)

	for _, resTx := range resTxs {
		resTx := resTx

		if _, ok := resBlocks[resTx.Height]; !ok {
			resBlock, err := node.Block(context.Background(), &resTx.Height)
			if err != nil {
				return nil, err
			}

			resBlocks[resTx.Height] = resBlock
		}
	}

	return resBlocks, nil
}

func mkTxResult(clientCtx client.Context, resTx *coretypes.ResultTx, resBlock *coretypes.ResultBlock) (*sdk.TxResponse, error) {
	decoder, err := decode.NewDecoder(decode.Options{
		SigningContext: clientCtx.TxConfig.SigningContext(),
		ProtoCodec:     clientCtx.Codec,
	})
	if err != nil {
		return nil, err
	}

	p, err := decoder.Decode(resTx.Tx)
	if err != nil {
		return nil, err
	}

	body := new(txtypes.TxBody)
	authInfo := new(txtypes.AuthInfo)

	err = clientCtx.Codec.Unmarshal(p.TxRaw.BodyBytes, body)
	if err != nil {
		return nil, err
	}

	err = clientCtx.Codec.Unmarshal(p.TxRaw.AuthInfoBytes, authInfo)
	if err != nil {
		return nil, err
	}

	tx := &txtypes.Tx{
		Body:       body,
		AuthInfo:   authInfo,
		Signatures: p.TxRaw.Signatures,
	}

	anyTx, err := codectypes.NewAnyWithValue(tx)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseResultTx(resTx, anyTx, resBlock.Block.Time.Format(time.RFC3339)), nil
}
