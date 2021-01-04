package rosetta

import (
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"

	tmcoretypes "github.com/tendermint/tendermint/rpc/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// tmResultTxsToSdkTxsWithHash converts tendermint result txs to cosmos sdk.Tx
func tmResultTxsToSdkTxsWithHash(decode sdk.TxDecoder, txs []*tmcoretypes.ResultTx) ([]*sdkTxWithHash, error) {
	converted := make([]*sdkTxWithHash, len(txs))
	for i, tx := range txs {
		sdkTx, err := decode(tx.Tx)
		if err != nil {
			return nil, err
		}
		converted[i] = &sdkTxWithHash{
			HexHash: fmt.Sprintf("%X", tx.Tx.Hash()),
			Code:    tx.TxResult.Code,
			Log:     tx.TxResult.Log,
			Tx:      sdkTx,
		}
	}

	return converted, nil
}

type sdkTxWithHash struct {
	HexHash string
	Code    uint32
	Log     string
	Tx      sdk.Tx
}

type PayloadReqMetadata struct {
	ChainID       string
	Sequence      uint64
	AccountNumber uint64
	Gas           uint64
	Memo          string
}

// getMetadataFromPayloadReq obtains the metadata from the request to /construction/payloads endpoint.
func getMetadataFromPayloadReq(req *types.ConstructionPayloadsRequest) (*PayloadReqMetadata, error) {
	chainID, ok := req.Metadata[OptionChainID].(string)
	if !ok {
		return nil, fmt.Errorf("chain_id metadata was not provided")
	}

	sequence, ok := req.Metadata[OptionSequence]
	if !ok {
		return nil, fmt.Errorf("sequence metadata was not provided")
	}

	seqNum, ok := sequence.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid sequence value")
	}

	accountNum, ok := req.Metadata[OptionAccountNumber]
	if !ok {
		return nil, fmt.Errorf("account_number metadata was not provided")
	}

	accNum, ok := accountNum.(float64)
	if !ok {
		fmt.Printf("this is type %T", accountNum)
		return nil, fmt.Errorf("invalid account_number value")
	}

	gasNum, ok := req.Metadata[OptionGas]
	if !ok {
		return nil, fmt.Errorf("gas metadata was not provided")
	}

	gasF64, ok := gasNum.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid gas value")
	}

	memo, ok := req.Metadata[OptionMemo]
	if !ok {
		memo = ""
	}

	memoStr, ok := memo.(string)
	if !ok {
		return nil, fmt.Errorf("invalid memo")
	}

	return &PayloadReqMetadata{
		ChainID:       chainID,
		Sequence:      uint64(seqNum),
		AccountNumber: uint64(accNum),
		Gas:           uint64(gasF64),
		Memo:          memoStr,
	}, nil
}
