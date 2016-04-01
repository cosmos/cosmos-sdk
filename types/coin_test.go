package types

import (
	"testing"
)

func TestCoins(t *testing.T) {
	coins := Coins{
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
		Coin{"TREE", 1},
	}

	if !coins.IsValid() {
		t.Fatal("Coins are valid")
	}

	if !coins.IsPositive() {
		t.Fatalf("Expected coins to be positive: %v", coins)
	}

	negCoins := coins.Negative()
	if negCoins.IsPositive() {
		t.Fatalf("Expected neg coins to not be positive: %v", negCoins)
	}

	sumCoins := coins.Plus(negCoins)
	if len(sumCoins) != 0 {
		t.Fatal("Expected 0 coins")
	}
}

func TestCoinsBadSort(t *testing.T) {
	coins := Coins{
		Coin{"TREE", 1},
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
	}

	if coins.IsValid() {
		t.Fatal("Coins are not sorted")
	}
}

func TestCoinsBadAmount(t *testing.T) {
	coins := Coins{
		Coin{"GAS", 1},
		Coin{"TREE", 0},
		Coin{"MINERAL", 1},
	}

	if coins.IsValid() {
		t.Fatal("Coins cannot include 0 amounts")
	}
}

func TestCoinsDuplicate(t *testing.T) {
	coins := Coins{
		Coin{"GAS", 1},
		Coin{"GAS", 1},
		Coin{"MINERAL", 1},
	}

	if coins.IsValid() {
		t.Fatal("Duplicate coin")
	}
}
