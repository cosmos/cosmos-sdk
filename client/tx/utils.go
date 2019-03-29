package tx

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

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

	info, err := formatTxResults(cdc, res.Txs)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// formatTxResults parses the indexed txs into a slice of TxResponse objects.
func formatTxResults(cdc *codec.Codec, res []*ctypes.ResultTx) ([]sdk.TxResponse, error) {
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

// ValidateTxResult performs transaction verification.
func ValidateTxResult(cliCtx context.CLIContext, res *ctypes.ResultTx) error {
	if !cliCtx.TrustNode {
		check, err := cliCtx.Verify(res.Height)
		if err != nil {
			return err
		}
		err = res.Proof.Validate(check.Header.DataHash)
		if err != nil {
			return err
		}
	}
	return nil
}

func formatTxResult(cdc *codec.Codec, res *ctypes.ResultTx) (sdk.TxResponse, error) {
	tx, err := parseTx(cdc, res.Tx)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	return sdk.NewResponseResultTx(res, tx), nil
}

func parseTx(cdc *codec.Codec, txBytes []byte) (sdk.Tx, error) {
	var tx auth.StdTx

	err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func queryTx(cdc *codec.Codec, cliCtx context.CLIContext, hashHexStr string) (sdk.TxResponse, error) {
	hash, err := hex.DecodeString(hashHexStr)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	node, err := cliCtx.GetNode()
	if err != nil {
		return sdk.TxResponse{}, err
	}

	res, err := node.Tx(hash, !cliCtx.TrustNode)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	if !cliCtx.TrustNode {
		if err = ValidateTxResult(cliCtx, res); err != nil {
			return sdk.TxResponse{}, err
		}
	}

	out, err := formatTxResult(cdc, res)
	if err != nil {
		return out, err
	}

	return out, nil
}
