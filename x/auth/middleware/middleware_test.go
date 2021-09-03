package middleware_test

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func generatePubKeysAndSignatures(n int, msg []byte, _ bool) (pubkeys []cryptotypes.PubKey, signatures [][]byte) {
	pubkeys = make([]cryptotypes.PubKey, n)
	signatures = make([][]byte, n)
	for i := 0; i < n; i++ {
		var privkey cryptotypes.PrivKey = secp256k1.GenPrivKey()

		// TODO: also generate ed25519 keys as below when ed25519 keys are
		//  actually supported, https://github.com/cosmos/cosmos-sdk/issues/4789
		// for now this fails:
		//if rand.Int63()%2 == 0 {
		//	privkey = ed25519.GenPrivKey()
		//} else {
		//	privkey = secp256k1.GenPrivKey()
		//}

		pubkeys[i] = privkey.PubKey()
		signatures[i], _ = privkey.Sign(msg)
	}
	return
}

func expectedGasCostByKeys(pubkeys []cryptotypes.PubKey) uint64 {
	cost := uint64(0)
	for _, pubkey := range pubkeys {
		pubkeyType := strings.ToLower(fmt.Sprintf("%T", pubkey))
		switch {
		case strings.Contains(pubkeyType, "ed25519"):
			cost += types.DefaultParams().SigVerifyCostED25519
		case strings.Contains(pubkeyType, "secp256k1"):
			cost += types.DefaultParams().SigVerifyCostSecp256k1
		default:
			panic("unexpected key type")
		}
	}
	return cost
}
