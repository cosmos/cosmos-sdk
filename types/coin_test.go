package types

import (
	"testing"

	cmn "github.com/tendermint/go-common"

	"github.com/stretchr/testify/assert"
)

func TestCoins(t *testing.T) {

	//Define the coins to be used in tests
	good := Coins{
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
		Coin{"TREE", 1},
	}
	neg := good.Negative()
	sum := good.Plus(neg)
	empty := Coins{
		Coin{"GOLD", 0},
	}
	badSort1 := Coins{
		Coin{"TREE", 1},
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		Coin{"GAS", 1},
		Coin{"TREE", 1},
		Coin{"MINERAL", 1},
	}
	badAmt := Coins{
		Coin{"GAS", 1},
		Coin{"TREE", 0},
		Coin{"MINERAL", 1},
	}
	dup := Coins{
		Coin{"GAS", 1},
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
	}

	//define the list of coin tests
	var testList = []struct {
		testPass bool
		errMsg   string
	}{
		{good.IsValid(), "Coins are valid"},
		{good.IsPositive(), cmn.Fmt("Expected coins to be positive: %v", good)},
		{good.IsGTE(empty), cmn.Fmt("Expected %v to be >= %v", good, empty)},
		{!neg.IsPositive(), cmn.Fmt("Expected neg coins to not be positive: %v", neg)},
		{len(sum) == 0, "Expected 0 coins"},
		{!badSort1.IsValid(), "Coins are not sorted"},
		{!badSort2.IsValid(), "Coins are not sorted"},
		{!badAmt.IsValid(), "Coins cannot include 0 amounts"},
		{!dup.IsValid(), "Duplicate coin"},
	}

	//execute the tests
	for _, tl := range testList {
		assert.True(t, tl.testPass, tl.errMsg)
	}
}
