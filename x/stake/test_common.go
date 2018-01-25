package stake

import (
	"encoding/hex"
	"fmt"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func newAddrs(n int) (addrs []crypto.Address) {
	for i := 0; i < n; i++ {
		addrs = append(addrs, []byte(fmt.Sprintf("addr%d", i)))
	}
	return
}

func newPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	//res, err = crypto.PubKeyFromBytes(pkBytes)
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd
}

// dummy pubkeys used for testing
var pks = []crypto.PubKey{
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB54"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB55"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB56"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB58"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB59"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB60"),
}

// NOTE: PubKey is supposed to be the binaryBytes of the crypto.PubKey
// instead this is just being set the address here for testing purposes
func candidatesFromActors(store sdk.KVStore, addrs []crypto.Address, amts []int) {
	for i := 0; i < len(addrs); i++ {
		c := &Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Owner:       addrs[i],
			Assets:      int64(amts[i]), //rational.New(amts[i]),
			Liabilities: int64(amts[i]), //rational.New(amts[i]),
			VotingPower: int64(amts[i]), //rational.New(amts[i]),
		}
		saveCandidate(store, c)
	}
}

func candidatesFromActorsEmpty(addrs []crypto.Address) (candidates Candidates) {
	for i := 0; i < len(addrs); i++ {
		c := &Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Owner:       addrs[i],
			Assets:      0, //rational.Zero,
			Liabilities: 0, //rational.Zero,
			VotingPower: 0, //rational.Zero,
		}
		candidates = append(candidates, c)
	}
	return
}

//// helper function test if Candidate is changed asabci.Validator
//func testChange(t *testing.T, val Validator, chg *abci.Validator) {
//assert := assert.New(t)
//assert.Equal(val.PubKey.Bytes(), chg.PubKey)
//assert.Equal(val.VotingPower.Evaluate(), chg.Power)
//}

//// helper function test if Candidate is removed as abci.Validator
//func testRemove(t *testing.T, val Validator, chg *abci.Validator) {
//assert := assert.New(t)
//assert.Equal(val.PubKey.Bytes(), chg.PubKey)
//assert.Equal(int64(0), chg.Power)
//}
