package types

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
)

func coinName(suffix int) string {
	return fmt.Sprintf("coinz%04d", suffix)
}

func BenchmarkCoinsAdditionIntersect(b *testing.B) {
	b.ReportAllocs()
	benchmarkingFunc := func(numCoinsA, numCoinsB int) func(b *testing.B) {
		return func(b *testing.B) {
			b.Helper()
			b.ReportAllocs()
			coinsA := Coins(make([]Coin, numCoinsA))
			coinsB := Coins(make([]Coin, numCoinsB))

			for i := range numCoinsA {
				coinsA[i] = NewCoin(coinName(i), math.NewInt(int64(i)))
			}
			for i := range numCoinsB {
				coinsB[i] = NewCoin(coinName(i), math.NewInt(int64(i)))
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				coinsA.Add(coinsB...)
			}
		}
	}

	benchmarkSizes := [][]int{{1, 1}, {5, 5}, {5, 20}, {1, 1000}, {2, 1000}}
	for i := range benchmarkSizes {
		sizeA := benchmarkSizes[i][0]
		sizeB := benchmarkSizes[i][1]
		b.Run(fmt.Sprintf("sizes: A_%d, B_%d", sizeA, sizeB), benchmarkingFunc(sizeA, sizeB))
	}
}

func BenchmarkCoinsAdditionNoIntersect(b *testing.B) {
	b.ReportAllocs()
	benchmarkingFunc := func(numCoinsA, numCoinsB int) func(b *testing.B) {
		return func(b *testing.B) {
			b.Helper()
			b.ReportAllocs()
			coinsA := Coins(make([]Coin, numCoinsA))
			coinsB := Coins(make([]Coin, numCoinsB))

			for i := range numCoinsA {
				coinsA[i] = NewCoin(coinName(numCoinsB+i), math.NewInt(int64(i)))
			}
			for i := range numCoinsB {
				coinsB[i] = NewCoin(coinName(i), math.NewInt(int64(i)))
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				coinsA.Add(coinsB...)
			}
		}
	}

	benchmarkSizes := [][]int{{1, 1}, {5, 5}, {5, 20}, {1, 1000}, {2, 1000}, {1000, 2}}
	for i := range benchmarkSizes {
		sizeA := benchmarkSizes[i][0]
		sizeB := benchmarkSizes[i][1]
		b.Run(fmt.Sprintf("sizes: A_%d, B_%d", sizeA, sizeB), benchmarkingFunc(sizeA, sizeB))
	}
}

func BenchmarkSumOfCoinAdds(b *testing.B) {
	// This benchmark tests the performance of adding a large number of coins
	// into a single coin set.
	// it does numAdds additions, each addition has (numIntersectingCoins) that contain denoms
	// already in the sum, and (coinsPerAdd - numIntersectingCoins) that are new denoms.
	benchmarkingFunc := func(numAdds, coinsPerAdd, numIntersectingCoins int, sumFn func([]Coins) Coins) func(b *testing.B) {
		return func(b *testing.B) {
			b.Helper()
			b.ReportAllocs()
			addCoins := make([]Coins, numAdds)
			nonIntersectingCoins := coinsPerAdd - numIntersectingCoins

			for i := range numAdds {
				intersectCoins := make([]Coin, numIntersectingCoins)
				num := math.NewInt(int64(i))
				for j := range numIntersectingCoins {
					intersectCoins[j] = NewCoin(coinName(j+1_000_000_000), num)
				}
				addCoins[i] = intersectCoins
				for j := range nonIntersectingCoins {
					addCoins[i] = addCoins[i].Add(NewCoin(coinName(i*nonIntersectingCoins+j), num))
				}
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				sumFn(addCoins)
			}
		}
	}

	MapCoinsSumFn := func(coins []Coins) Coins {
		sum := MapCoins{}
		for _, coin := range coins {
			sum.Add(coin...)
		}
		return sum.ToCoins()
	}
	CoinsSumFn := func(coins []Coins) Coins {
		sum := Coins{}
		for _, coin := range coins {
			sum = sum.Add(coin...)
		}
		return sum
	}

	// larger benchmarks with non-overlapping coins won't terminate in reasonable timeframes with sdk.Coins
	// they work fine with MapCoins
	benchmarkSizes := [][]int{{5, 2, 1000}, {10, 10, 10000}}
	sumFns := []struct {
		name string
		fn   func([]Coins) Coins
	}{
		{"MapCoins", MapCoinsSumFn}, {"Coins", CoinsSumFn},
	}
	for i := range benchmarkSizes {
		for j := range 2 {
			coinsPerAdd := benchmarkSizes[i][0]
			intersectingCoinsPerAdd := benchmarkSizes[i][1]
			numAdds := benchmarkSizes[i][2]
			sumFn := sumFns[j]
			b.Run(fmt.Sprintf("Fn: %s, num adds: %d, coinsPerAdd: %d, intersecting: %d",
				sumFn.name, numAdds, coinsPerAdd, intersectingCoinsPerAdd),
				benchmarkingFunc(numAdds, coinsPerAdd, intersectingCoinsPerAdd, sumFn.fn))
		}
	}
}
