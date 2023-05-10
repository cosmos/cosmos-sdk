package aminocompat

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

var sink any

var coinsPos = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
var addr = sdk.AccAddress("addr1")

func BenchmarkMsgDepositGetSignBytes(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg := govtypes.NewMsgDeposit(addr, 0, coinsPos)
		sink = msg.GetSignBytes()
	}

	if sink == nil {
		b.Fatal("Benchmark did not run")
	}
	sink = nil
}

func BenchmarkMsgDepositAllClearThenJSONMarshal(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg := govtypes.NewMsgDeposit(addr, 0, coinsPos)
		if err := AllClear(msg); err != nil {
			b.Fatal(err)
		}
		// Straight away go to JSON marshalling.
		blob, err := json.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
		sink = blob
	}

	if sink == nil {
		b.Fatal("Benchmark did not run")
	}
	sink = nil
}
