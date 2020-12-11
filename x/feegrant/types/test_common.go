// noalias

package types

// import (
// 	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
// )

// func NewTestTx(ctx sdk.Context, msgs []sdk.Msg, privs []cryptotypes.PrivKey, accNums []uint64, seqs []uint64, fee GrantedFee) sdk.Tx {
// 	sigs := make([]legacytx.StdSignature, len(privs))
// 	for i, priv := range privs {
// 		signBytes := StdSignBytes(ctx.ChainID(), accNums[i], seqs[i], fee, msgs, "")

// 		sig, err := priv.Sign(signBytes)
// 		if err != nil {
// 			panic(err)
// 		}

// 		sigs[i] = legacytx.StdSignature{PubKey: priv.PubKey(), Signature: sig}
// 	}

// 	tx := NewFeeGrantTx(msgs, fee, sigs, "")
// 	return tx
// }
