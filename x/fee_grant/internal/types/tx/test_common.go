// noalias
package tx

import (
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func NewTestTx(ctx sdk.Context, msgs []sdk.Msg, privs []crypto.PrivKey, accNums []uint64, seqs []uint64, fee GrantedFee) sdk.Tx {
	sigs := make([]authtypes.StdSignature, len(privs))
	for i, priv := range privs {
		signBytes := StdSignBytes(ctx.ChainID(), accNums[i], seqs[i], fee, msgs, "")

		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = authtypes.StdSignature{PubKey: priv.PubKey(), Signature: sig}
	}

	tx := NewFeeGrantTx(msgs, fee, sigs, "")
	return tx
}
