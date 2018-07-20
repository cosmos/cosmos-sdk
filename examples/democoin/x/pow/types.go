package pow

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgMine - mine some coins with PoW
type MsgMine struct {
	Sender     sdk.AccAddress `json:"sender"`
	Difficulty uint64         `json:"difficulty"`
	Count      uint64         `json:"count"`
	Nonce      uint64         `json:"nonce"`
	Proof      []byte         `json:"proof"`
}

// enforce the msg type at compile time
var _ sdk.Msg = MsgMine{}

// NewMsgMine - construct mine message
func NewMsgMine(sender sdk.AccAddress, difficulty uint64, count uint64, nonce uint64, proof []byte) MsgMine {
	return MsgMine{sender, difficulty, count, nonce, proof}
}

// nolint
func (msg MsgMine) Type() string                 { return "pow" }
func (msg MsgMine) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }
func (msg MsgMine) String() string {
	return fmt.Sprintf("MsgMine{Sender: %s, Difficulty: %d, Count: %d, Nonce: %d, Proof: %s}", msg.Sender, msg.Difficulty, msg.Count, msg.Nonce, msg.Proof)
}

// validate the mine message
func (msg MsgMine) ValidateBasic() sdk.Error {
	// check hash
	var data []byte
	// hash must include sender, so no other users can race the tx
	data = append(data, []byte(msg.Sender)...)
	countBytes := strconv.FormatUint(msg.Count, 16)
	// hash must include count so proof-of-work solutions cannot be replayed
	data = append(data, countBytes...)
	nonceBytes := strconv.FormatUint(msg.Nonce, 16)
	data = append(data, nonceBytes...)
	hash := crypto.Sha256(data)
	hashHex := make([]byte, hex.EncodedLen(len(hash)))
	hex.Encode(hashHex, hash)
	hashHex = hashHex[:16]
	if !bytes.Equal(hashHex, msg.Proof) {
		return ErrInvalidProof(DefaultCodespace, fmt.Sprintf("hashHex: %s, proof: %s", hashHex, msg.Proof))
	}

	// check proof below difficulty
	// difficulty is linear - 1 = all hashes, 2 = half of hashes, 3 = third of hashes, etc
	target := math.MaxUint64 / msg.Difficulty
	hashUint, err := strconv.ParseUint(string(msg.Proof), 16, 64)
	if err != nil {
		return ErrInvalidProof(DefaultCodespace, fmt.Sprintf("proof: %s", msg.Proof))
	}
	if hashUint >= target {
		return ErrNotBelowTarget(DefaultCodespace, fmt.Sprintf("hashuint: %d, target: %d", hashUint, target))
	}

	return nil
}

// get the mine message sign bytes
func (msg MsgMine) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}
