package pow

import (
	"encoding/hex"
	"math"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

func GenerateMineMsg(sender sdk.Address, count uint64, difficulty uint64) MineMsg {
	nonce, hash := mine(sender, count, difficulty)
	return NewMineMsg(sender, difficulty, count, nonce, hash)
}

func hash(sender sdk.Address, count uint64, nonce uint64) []byte {
	var bytes []byte
	bytes = append(bytes, []byte(sender)...)
	countBytes := strconv.FormatUint(count, 16)
	bytes = append(bytes, countBytes...)
	nonceBytes := strconv.FormatUint(nonce, 16)
	bytes = append(bytes, nonceBytes...)
	hash := crypto.Sha256(bytes)
	// uint64, so we just use the first 8 bytes of the hash
	// this limits the range of possible difficulty values (as compared to uint256), but fine for proof-of-concept
	ret := make([]byte, hex.EncodedLen(len(hash)))
	hex.Encode(ret, hash)
	return ret[:16]
}

func mine(sender sdk.Address, count uint64, difficulty uint64) (uint64, []byte) {
	target := math.MaxUint64 / difficulty
	for nonce := uint64(0); ; nonce++ {
		hash := hash(sender, count, nonce)
		hashuint, err := strconv.ParseUint(string(hash), 16, 64)
		if err != nil {
			panic(err)
		}
		if hashuint < target {
			return nonce, hash
		}
	}
}
