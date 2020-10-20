package rosetta

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/tendermint/cosmos-rosetta-gateway/rosetta"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func (l launchpad) ConstructionHash(ctx context.Context, req *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	bz, err := hex.DecodeString(req.SignedTransaction)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, "error decoding tx")
	}

	var stdTx auth.StdTx
	err = Codec.UnmarshalJSON(bz, &stdTx)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, "invalid tx")
	}

	txBytes, err := Codec.MarshalBinaryLengthPrefixed(stdTx)
	if err != nil {
		return nil, rosetta.WrapError(ErrInvalidTransaction, "invalid tx")
	}

	hash := sha256.Sum256(txBytes)
	bzHash := hash[:]

	hashString := hex.EncodeToString(bzHash)

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: strings.ToUpper(hashString),
		},
	}, nil
}
