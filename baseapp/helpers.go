package baseapp

import (
	"bytes"
	`fmt`
	"regexp"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// nolint - Mostly for testing
func (app *BaseApp) Check(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(RunTxModeCheck, nil, tx)
}

// nolint - full tx execution
func (app *BaseApp) Simulate(txBytes []byte, tx sdk.Tx) (result sdk.Result) {
	return app.runTx(RunTxModeSimulate, txBytes, tx)
}

// nolint
func (app *BaseApp) Deliver(tx sdk.Tx) (result sdk.Result) {
	return app.runTx(RunTxModeDeliver, nil, tx)
}

// Context with current {check, deliver}State of the app
// used by tests
func (app *BaseApp) NewContext(isCheckTx bool, header abci.Header) sdk.Context {
	if isCheckTx {
		return sdk.NewContext(app.checkState.ms, header, true, app.logger).
			WithMinGasPrices(app.minGasPrices)
	}

	return sdk.NewContext(app.deliverState.ms, header, false, app.logger)
}

func getFeeFromTags(res sdk.Result) (i int, fee sdk.Coins) {
	for i, tag := range res.Tags.ToKVPairs() {
		if bytes.EqualFold(tag.Key, []byte(sdk.Fee_TagName)) {
			//fmt.Printf("%s: %s\n", string(tag.Key), string(tag.Value))
			//res.Tags = append(res.Tags[0:i], res.Tags[i+1:]...)
			return i, strToCoins(string(tag.Value))
		}
	}
	return i, sdk.Coins{}
}

func strToCoins(amount string) sdk.Coins {
	var res sdk.Coins
	coinStrs := strings.Split(amount, ",")
	for _, coinStr := range coinStrs {
		coin := strings.Split(coinStr, ":")
		if len(coin) == 2 {
			var c sdk.Coin
			c.Denom = coin[1]
			coinDec := sdk.MustNewDecFromStr(coin[0])
			c.Amount = sdk.NewIntFromBigInt(coinDec.Int)
			res = append(res, c)
		}
	}
	return res
}

func coins2str(coins sdk.Coins)string{
	if len(coins) == 0 {
		return ""
	}

	out := ""
	for _, coin := range coins {
		out += fmt.Sprintf("%v,", coin2str(coin))
	}
	return out[:len(out)-1]
}

// String provides a human-readable representation of a coin
func coin2str(coin sdk.Coin) string {
	dec := sdk.NewDecFromIntWithPrec(coin.Amount, sdk.Precision)
	return fmt.Sprintf("%s %v", dec, coin.Denom)
}


