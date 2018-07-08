package pow

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewMsgMine(t *testing.T) {
	addr := sdk.AccAddress([]byte("sender"))
	msg := MsgMine{addr, 0, 0, 0, []byte("")}
	equiv := NewMsgMine(addr, 0, 0, 0, []byte(""))
	require.Equal(t, msg, equiv, "%s != %s", msg, equiv)
}

func TestMsgMineType(t *testing.T) {
	addr := sdk.AccAddress([]byte("sender"))
	msg := MsgMine{addr, 0, 0, 0, []byte("")}
	require.Equal(t, msg.Type(), "pow")
}

func TestMsgMineValidation(t *testing.T) {
	addr := sdk.AccAddress([]byte("sender"))
	otherAddr := sdk.AccAddress([]byte("another"))
	count := uint64(0)

	for difficulty := uint64(1); difficulty < 1000; difficulty += 100 {

		count++
		nonce, proof := mine(addr, count, difficulty)
		msg := MsgMine{addr, difficulty, count, nonce, proof}
		err := msg.ValidateBasic()
		require.Nil(t, err, "error with difficulty %d - %+v", difficulty, err)

		msg.Count++
		err = msg.ValidateBasic()
		require.NotNil(t, err, "count was wrong, should have thrown error with msg %s", msg)

		msg.Count--
		msg.Nonce++
		err = msg.ValidateBasic()
		require.NotNil(t, err, "nonce was wrong, should have thrown error with msg %s", msg)

		msg.Nonce--
		msg.Sender = otherAddr
		err = msg.ValidateBasic()
		require.NotNil(t, err, "sender was wrong, should have thrown error with msg %s", msg)
	}
}

func TestMsgMineString(t *testing.T) {
	addr := sdk.AccAddress([]byte("sender"))
	msg := MsgMine{addr, 0, 0, 0, []byte("abc")}
	res := msg.String()
	require.Equal(t, res, "MsgMine{Sender: cosmosaccaddr1wdjkuer9wg4wml9c, Difficulty: 0, Count: 0, Nonce: 0, Proof: abc}")
}

func TestMsgMineGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress([]byte("sender"))
	msg := MsgMine{addr, 1, 1, 1, []byte("abc")}
	res := msg.GetSignBytes()
	require.Equal(t, string(res), `{"count":1,"difficulty":1,"nonce":1,"proof":"YWJj","sender":"cosmosaccaddr1wdjkuer9wg4wml9c"}`)
}

func TestMsgMineGetSigners(t *testing.T) {
	addr := sdk.AccAddress([]byte("sender"))
	msg := MsgMine{addr, 1, 1, 1, []byte("abc")}
	res := msg.GetSigners()
	require.Equal(t, fmt.Sprintf("%v", res), "[73656E646572]")
}
