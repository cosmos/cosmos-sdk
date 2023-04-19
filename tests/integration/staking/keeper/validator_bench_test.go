package keeper_test

import "testing"

func BenchmarkGetValidator(b *testing.B) {
	// 900 is the max number we are allowed to use in order to avoid simtestutil.CreateTestPubKeys
	// panic: encoding/hex: odd length hex string
	powersNumber := 900

	var totalPower int64
	powers := make([]int64, powersNumber)
	for i := range powers {
		powers[i] = int64(i)
		totalPower += int64(i)
	}

	f, _, valAddrs, vals := initValidators(b, totalPower, len(powers), powers)

	for _, validator := range vals {
		f.stakingKeeper.SetValidator(f.sdkCtx, validator)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, addr := range valAddrs {
			_, _ = f.stakingKeeper.GetValidator(f.sdkCtx, addr)
		}
	}
}
