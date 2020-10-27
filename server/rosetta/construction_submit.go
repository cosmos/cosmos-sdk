package rosetta

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"
)

type BroadcastReq struct {
	Tx   json.RawMessage `json:"tx"`
	Mode string          `json:"mode"`
}

func (l launchpad) ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	if l.properties.OfflineMode {
		return nil, ErrEndpointDisabledOfflineMode
	}

	bz, err := hex.DecodeString(req.SignedTransaction)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, "error decoding tx")
	}

	var tx map[string]json.RawMessage
	err = json.Unmarshal(bz, &tx)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, "error unmarshaling tx")
	}

	bReq := BroadcastReq{
		Tx:   tx["value"],
		Mode: "block",
	}

	bytes, err := json.Marshal(bReq)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, "error decoding tx")
	}

	resp, err := l.cosmos.PostTx(ctx, bytes)
	if err != nil {
		return nil, rosetta.WrapError(ErrNodeConnection, fmt.Sprintf("error broadcasting tx: %s", err))
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: resp.TxHash,
		},
		Metadata: map[string]interface{}{
			"log": resp.RawLog,
		},
	}, nil
}
