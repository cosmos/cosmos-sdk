package conversion

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/tendermint/btcd/btcec"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// RosettaSignatureToCosmos converts a rosetta signature to a cosmos one
func RosettaSignatureToCosmos(sig *types.Signature) (signing.SignatureV2, error) {
	if sig.SignatureType != types.Ecdsa {
		return signing.SignatureV2{}, fmt.Errorf("bad signature type: %s, expected: %s", sig.SignatureType, types.Ecdsa)
	}
	if sig.PublicKey.CurveType != types.Secp256k1 {
		return signing.SignatureV2{}, fmt.Errorf("bad signature curve: %s, expected: %s", sig.PublicKey.CurveType, types.Secp256k1)
	}
	// get public key
	_, err := btcec.ParsePubKey(sig.PublicKey.Bytes, btcec.S256())
	if err != nil {
		return signing.SignatureV2{}, fmt.Errorf("unable to parse public key: %s", err.Error())
	}
	panic("not implemented :(")
}

func GetFeeOpFromCoins(coins sdk.Coins, account string) []*types.Operation {
	var feeOps []*types.Operation
	for _, coin := range coins {
		op := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(0),
			},
			Type: rosetta.OperationFee,
			Account: &types.AccountIdentifier{
				Address: account,
			},
			Amount: &types.Amount{
				Value: coin.Amount.String(),
				Currency: &types.Currency{
					Symbol: coin.Denom,
				},
			},
		}
		feeOps = append(feeOps, op)
	}
	return feeOps
}
