package auth

// import (
// 	"fmt"
// 	"testing"

// 	crypto "github.com/tendermint/go-crypto"
// 	cmn "github.com/tendermint/tmlibs/common"

// 	sdk "github.com/cosmos/cosmos-sdk"
// 	"github.com/cosmos/cosmos-sdk/state"
// 	"github.com/cosmos/cosmos-sdk/util"
// )

// func makeSignTx() sdk.Tx {
// 	key := crypto.GenPrivKeyEd25519().Wrap()
// 	payload := cmn.RandBytes(32)
// 	tx := NewSig(util.RawTx{payload})
// 	Sign(tx, key)
// 	return tx.Wrap()
// }

// func makeMultiSignTx(cnt int) sdk.Tx {
// 	payload := cmn.RandBytes(32)
// 	tx := NewMulti(util.RawTx{payload})
// 	for i := 0; i < cnt; i++ {
// 		key := crypto.GenPrivKeyEd25519().Wrap()
// 		Sign(tx, key)
// 	}
// 	return tx.Wrap()
// }

// func makeHandler() sdk.Handler {
// 	return sdk.ChainDecorators(
// 		Signatures{},
// 	).WithHandler(
// 		util.OKHandler{},
// 	)
// }

// func BenchmarkCheckOneSig(b *testing.B) {
// 	tx := makeSignTx()
// 	h := makeHandler()
// 	store := state.NewMemKVStore()
// 	for i := 1; i <= b.N; i++ {
// 		ctx := util.MockContext("foo", 100)
// 		_, err := h.DeliverTx(ctx, store, tx)
// 		// never should error
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// }

// func BenchmarkCheckMultiSig(b *testing.B) {
// 	sigs := []int{1, 3, 8, 20}
// 	for _, cnt := range sigs {
// 		label := fmt.Sprintf("%dsigs", cnt)
// 		b.Run(label, func(sub *testing.B) {
// 			benchmarkCheckMultiSig(sub, cnt)
// 		})
// 	}
// }

// func benchmarkCheckMultiSig(b *testing.B, cnt int) {
// 	tx := makeMultiSignTx(cnt)
// 	h := makeHandler()
// 	store := state.NewMemKVStore()
// 	for i := 1; i <= b.N; i++ {
// 		ctx := util.MockContext("foo", 100)
// 		_, err := h.DeliverTx(ctx, store, tx)
// 		// never should error
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// }
