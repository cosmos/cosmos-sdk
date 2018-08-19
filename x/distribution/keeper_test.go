package stake

//// test if is a gotValidator from the last update
//func TestGetTotalPrecommitVotingPower(t *testing.T) {
//ctx, _, keeper := createTestInput(t, false, 0)

//amts := []int64{10000, 1000, 100, 10, 1}
//var candidatesIn [5]Candidate
//for i, amt := range amts {
//candidatesIn[i] = NewCandidate(addrVals[i], pks[i], Description{})
//candidatesIn[i].BondedShares = sdk.NewRat(amt)
//candidatesIn[i].DelegatorShares = sdk.NewRat(amt)
//keeper.setCandidate(ctx, candidatesIn[i])
//}

//// test that an empty gotValidator set doesn't have any gotValidators
//gotValidators := keeper.GetValidators(ctx)
//require.Equal(t, 5, len(gotValidators))

//totPow := keeper.GetTotalPrecommitVotingPower(ctx)
//exp := sdk.NewRat(11111)
//require.True(t, exp.Equal(totPow), "exp %v, got %v", exp, totPow)

//// set absent gotValidators to be the 1st and 3rd record sorted by pubKey address
//ctx = ctx.WithAbsentValidators([]int32{1, 3})
//totPow = keeper.GetTotalPrecommitVotingPower(ctx)

//// XXX verify that this order should infact exclude these two records
//exp = sdk.NewRat(11100)
//require.True(t, exp.Equal(totPow), "exp %v, got %v", exp, totPow)
//}
