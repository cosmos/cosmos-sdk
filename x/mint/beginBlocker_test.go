package mint

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestBeginBlocker(t *testing.T) {
	mintParams := Params{
		MintDenom:     sdk.DefaultBondDenom,
		InflationRate: sdk.NewDecWithPrec(1, 2),
		BlocksPerYear: uint64(100),
	}
	var balance int64 = 10000
	mapp, _ := getMockApp(t, 1, balance, mintParams)

	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: int64(2)}})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{}).WithBlockHeight(int64(2))

	// mint rate test
	minter := mapp.mintKeeper.GetMinterCustom(ctx)
	ratePerBlock0 := minter.MintedPerBlock.AmountOf(sdk.DefaultBondDenom)

	annualProvisions := mintParams.InflationRate.Mul(sdk.NewDec(balance))
	provisionAmtPerBlock := annualProvisions.Quo(sdk.NewDec(int64(mintParams.BlocksPerYear)))
	assert.EqualValues(t, ratePerBlock0, provisionAmtPerBlock)

	var curHeight int64 = 2
	for ; curHeight < 101; curHeight++ {
		mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: curHeight}})
		mapp.EndBlock(abci.RequestEndBlock{Height: curHeight})
		mapp.Commit()
	}

	// this year mint test
	curSupply0 := mapp.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(sdk.DefaultBondDenom)
	curCoin0 := curSupply0
	rawCoin := sdk.NewDec(balance)
	assert.EqualValues(t, curCoin0.Sub(rawCoin), ratePerBlock0.Mul(sdk.NewDec(curHeight-1)))

	// next year mint test
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: curHeight}})
	ctx = mapp.BaseApp.NewContext(false, abci.Header{}).WithBlockHeight(curHeight)
	curSupply1 := mapp.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(sdk.DefaultBondDenom)
	curCoin1 := curSupply1

	minter = mapp.mintKeeper.GetMinterCustom(ctx)
	ratePerBlock1 := minter.MintedPerBlock.AmountOf(sdk.DefaultBondDenom)
	lastMintTotalSupply := mintParams.InflationRate.Mul(sdk.NewDec(balance)).Add(sdk.NewDec(balance))
	annualProvisions = mintParams.InflationRate.Mul(lastMintTotalSupply)
	provisionAmtPerBlock = annualProvisions.Quo(sdk.NewDec(int64(mintParams.BlocksPerYear)))
	assert.EqualValues(t, ratePerBlock1, provisionAmtPerBlock)

	// annual mint test
	step1Mint := ratePerBlock0.Mul(sdk.NewDec(100))
	step2Mint := ratePerBlock1.Mul(sdk.NewDec(curHeight - 100))
	totalMint := step1Mint.Add(step2Mint)
	assert.EqualValues(t, curCoin1.Sub(rawCoin), totalMint)
}
