package hash

import (
	"crypto/sha256"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/gogoproto/proto"
)

var addressSize = 20

// AddressBasic
// ADR-28 compatible basic address
// hash(hash(typ) + key)[:20]
func AddressBasic(pubKey types.PubKey) []byte {
	msgType := proto.MessageName(pubKey)
	msgTypeHash := sha256.Sum256([]byte(msgType))                 // hash(typ)
	preHashedAddress := append(msgTypeHash[:], pubKey.Bytes()...) // hash(typ) + key
	hashedAddress := sha256.Sum256(preHashedAddress)
	return hashedAddress[:addressSize] // hash(hash(typ) + key)[:A_LEN]
}
