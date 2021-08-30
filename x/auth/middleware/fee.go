package middleware

// import (
// 	"github.com/cosmos/cosmos-sdk/types/tx"
// 	"github.com/cosmos/cosmos-sdk/x/auth/types"
// )

// type deductFeeMiddleware struct {
// 	ak             AccountKeeper
// 	bankKeeper     types.BankKeeper
// 	feegrantKeeper FeegrantKeeper
// 	next           tx.Handler
// }

// func DeductFeeMiddleware(txh tx.Handler) tx.Handler {
// 	return mempoolFeeMiddleware{
// 		next: txh,
// 	}
// }
