package types

import (
	"fmt"
	"testing"
)

func BenchmarkCoinsAdditionIntersect(b *testing.B) {
	benchmarkingFunc := func(numCoinsA int, numCoinsB int) func(b *testing.B) {
		return func(b *testing.B) {
			coinsA := Coins(make([]Coin, numCoinsA))
			coinsB := Coins(make([]Coin, numCoinsB))

			for i := 0; i < numCoinsA; i++ {
				coinsA[i] = NewCoin("COINZ_"+string(i), NewInt(int64(i)))
			}
			for i := 0; i < numCoinsB; i++ {
				coinsB[i] = NewCoin("COINZ_"+string(i), NewInt(int64(i)))
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				coinsA.Plus(coinsB)
			}
		}
	}

	benchmarkSizes := [][]int{{1, 1}, {5, 5}, {5, 20}, {1, 1000}, {2, 1000}}
	for i := 0; i < len(benchmarkSizes); i++ {
		sizeA := benchmarkSizes[i][0]
		sizeB := benchmarkSizes[i][1]
		b.Run(fmt.Sprintf("sizes: A_%d, B_%d", sizeA, sizeB), benchmarkingFunc(sizeA, sizeB))
	}
}

func BenchmarkCoinsAdditionNoIntersect(b *testing.B) {
	benchmarkingFunc := func(numCoinsA int, numCoinsB int) func(b *testing.B) {
		return func(b *testing.B) {
			coinsA := Coins(make([]Coin, numCoinsA))
			coinsB := Coins(make([]Coin, numCoinsB))

			for i := 0; i < numCoinsA; i++ {
				coinsA[i] = NewCoin("COINZ_"+string(numCoinsB+i), NewInt(int64(i)))
			}
			for i := 0; i < numCoinsB; i++ {
				coinsB[i] = NewCoin("COINZ_"+string(i), NewInt(int64(i)))
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				coinsA.Plus(coinsB)
			}
		}
	}

	benchmarkSizes := [][]int{{1, 1}, {5, 5}, {5, 20}, {1, 1000}, {2, 1000}, {1000, 2}}
	for i := 0; i < len(benchmarkSizes); i++ {
		sizeA := benchmarkSizes[i][0]
		sizeB := benchmarkSizes[i][1]
		b.Run(fmt.Sprintf("sizes: A_%d, B_%d", sizeA, sizeB), benchmarkingFunc(sizeA, sizeB))
	}
}
