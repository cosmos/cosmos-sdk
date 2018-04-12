package pow

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewMineMsg(t *testing.T) {
	addr := sdk.Address([]byte("sender"))
	msg := MineMsg{addr, 0, 0, 0, []byte("")}
	equiv := NewMineMsg(addr, 0, 0, 0, []byte(""))
	assert.Equal(t, msg, equiv, "%s != %s", msg, equiv)
}

func TestMineMsgType(t *testing.T) {
	addr := sdk.Address([]byte("sender"))
	msg := MineMsg{addr, 0, 0, 0, []byte("")}
	assert.Equal(t, msg.Type(), "pow")
}

func TestMineMsgValidation(t *testing.T) {
	addr := sdk.Address([]byte("sender"))
	otherAddr := sdk.Address([]byte("another"))
	count := uint64(0)
	for difficulty := uint64(1); difficulty < 1000; difficulty += 100 {
		count += 1
		nonce, proof := mine(addr, count, difficulty)
		msg := MineMsg{addr, difficulty, count, nonce, proof}
		err := msg.ValidateBasic()
		assert.Nil(t, err, "error with difficulty %d - %+v", difficulty, err)

		msg.Count += 1
		err = msg.ValidateBasic()
		assert.NotNil(t, err, "count was wrong, should have thrown error with msg %s", msg)

		msg.Count -= 1
		msg.Nonce += 1
		err = msg.ValidateBasic()
		assert.NotNil(t, err, "nonce was wrong, should have thrown error with msg %s", msg)

		msg.Nonce -= 1
		msg.Sender = otherAddr
		err = msg.ValidateBasic()
		assert.NotNil(t, err, "sender was wrong, should have thrown error with msg %s", msg)
	}
}

func TestMineMsgString(t *testing.T) {
	addr := sdk.Address([]byte("sender"))
	msg := MineMsg{addr, 0, 0, 0, []byte("abc")}
	res := msg.String()
	assert.Equal(t, res, "MineMsg{Sender: 73656E646572, Difficulty: 0, Count: 0, Nonce: 0, Proof: abc}")
}

func TestMineMsgGet(t *testing.T) {
	addr := sdk.Address([]byte("sender"))
	msg := MineMsg{addr, 0, 0, 0, []byte("")}
	res := msg.Get(nil)
	assert.Nil(t, res)
}

func TestMineMsgGetSignBytes(t *testing.T) {
	addr := sdk.Address([]byte("sender"))
	msg := MineMsg{addr, 1, 1, 1, []byte("abc")}
	res := msg.GetSignBytes()
	assert.Equal(t, string(res), `{"sender":"73656E646572","difficulty":1,"count":1,"nonce":1,"proof":"YWJj"}`)
}

func TestMineMsgGetSigners(t *testing.T) {
	addr := sdk.Address([]byte("sender"))
	msg := MineMsg{addr, 1, 1, 1, []byte("abc")}
	res := msg.GetSigners()
	assert.Equal(t, fmt.Sprintf("%v", res), "[73656E646572]")
}
