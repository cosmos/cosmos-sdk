package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewMsgChangeKey(t *testing.T) {}

func TestMsgChangeKeyType(t *testing.T) {
	addr1 := sdk.Address([]byte("input"))
	newPubKey := crypto.GenPrivKeyEd25519().PubKey()

	var msg = MsgChangeKey{
		Address:   addr1,
		NewPubKey: newPubKey,
	}

	assert.Equal(t, msg.Type(), "auth")
}

func TestMsgChangeKeyValidation(t *testing.T) {

	addr1 := sdk.Address([]byte("input"))

	// emptyPubKey := crypto.PubKeyEd25519{}
	// var msg = MsgChangeKey{
	// 	Address:   addr1,
	// 	NewPubKey: emptyPubKey,
	// }

	// // fmt.Println(msg.NewPubKey.Empty())
	// fmt.Println(msg.NewPubKey.Bytes())

	// assert.NotNil(t, msg.ValidateBasic())

	newPubKey := crypto.GenPrivKeyEd25519().PubKey()
	msg := MsgChangeKey{
		Address:   addr1,
		NewPubKey: newPubKey,
	}
	assert.Nil(t, msg.ValidateBasic())
}
