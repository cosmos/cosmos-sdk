package construction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Client struct {
	txDec sdk.TxDecoder
}

//
//func (c *Client) MakeSignature(hexTx string, signers []interface{}) (tx sdk.Tx, err error) {
//	bz, err := hex.DecodeString(hexTx)
//	if err != nil {
//		return nil, rosetta.WrapError(rosetta.ErrBadArgument, "invalid tx hex string: "+err.Error())
//	}
//	x, err := c.txDec(bz)
//	if err != nil {
//
//	}
//}
