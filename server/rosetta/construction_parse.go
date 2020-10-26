package rosetta

import (
	"context"
	"encoding/hex"

	"github.com/coinbase/rosetta-sdk-go/types"

	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// ConstructionParse implements the /construction/parse endpoint.
func (l launchpad) ConstructionParse(ctx context.Context, request *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	rawTx, err := hex.DecodeString(request.Transaction)
	if err != nil {
		return nil, ErrInvalidTransaction
	}

	var stdTx auth.StdTx
	err = l.cdc.UnmarshalJSON(rawTx, &stdTx)
	if err != nil {
		return nil, ErrInvalidTransaction
	}

	signers := make([]string, len(stdTx.Signatures))
	for i, sig := range stdTx.Signatures {
		addr, err := cosmostypes.AccAddressFromHex(sig.PubKey.Address().String())
		if err != nil {
			return nil, ErrInvalidTransaction
		}
		signers[i] = addr.String()
	}

	return &types.ConstructionParseResponse{
		Operations: toOperations(stdTx.Msgs, false, true),
		Signers:    signers,
	}, nil
}
