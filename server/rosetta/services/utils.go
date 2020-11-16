package services

import (
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/cosmos/cosmos-sdk/server/rosetta"
)

type PayloadReqMetadata struct {
	ChainID       string
	Sequence      uint64
	AccountNumber uint64
	Gas           uint64
	Memo          string
}

// GetMetadataFromPayloadReq obtains the metadata from the request to /construction/payloads endpoint.
func GetMetadataFromPayloadReq(req *types.ConstructionPayloadsRequest) (*PayloadReqMetadata, error) {
	chainID, ok := req.Metadata[rosetta.ChainID].(string)
	if !ok {
		return nil, fmt.Errorf("chain_id metadata was not provided")
	}

	sequence, ok := req.Metadata[rosetta.Sequence]
	if !ok {
		return nil, fmt.Errorf("sequence metadata was not provided")
	}
	seqNum, ok := sequence.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid sequence value")
	}

	accountNum, ok := req.Metadata[rosetta.AccountNumber]
	if !ok {
		return nil, fmt.Errorf("account_number metadata was not provided")
	}
	accNum, ok := accountNum.(float64)
	if !ok {
		fmt.Printf("this is type %T", accountNum)
		return nil, fmt.Errorf("invalid account_number value")
	}

	gasNum, ok := req.Metadata[rosetta.OptionGas]
	if !ok {
		return nil, fmt.Errorf("gas metadata was not provided")
	}
	gasF64, ok := gasNum.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid gas value")
	}

	memo, ok := req.Metadata[rosetta.OptionMemo]
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
