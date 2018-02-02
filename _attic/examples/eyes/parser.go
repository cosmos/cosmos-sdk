package eyes

import sdk "github.com/cosmos/cosmos-sdk"

// Parser converts bytes into a tx struct
type Parser struct{}

var _ sdk.Decorator = Parser{}

// CheckTx makes sure we are on the proper chain
// - fulfills Decorator interface
func (c Parser) CheckTx(ctx sdk.Context, store sdk.SimpleDB,
	txBytes interface{}, next sdk.Checker) (res sdk.CheckResult, err error) {

	tx, err := LoadTx(txBytes.([]byte))
	if err != nil {
		return res, err
	}
	return next.CheckTx(ctx, store, tx)
}

// DeliverTx makes sure we are on the proper chain
// - fulfills Decorator interface
func (c Parser) DeliverTx(ctx sdk.Context, store sdk.SimpleDB,
	txBytes interface{}, next sdk.Deliverer) (res sdk.DeliverResult, err error) {

	tx, err := LoadTx(txBytes.([]byte))
	if err != nil {
		return res, err
	}
	return next.DeliverTx(ctx, store, tx)
}
