package slashing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgUnrevokeGetSignBytes(t *testing.T) {
	addr := sdk.Address("abcd")
	msg := NewMsgUnrevoke(addr)
	bytes := msg.GetSignBytes()
	assert.Equal(t, string(bytes), `{"address":"cosmosvaladdr1v93xxeqamr0mv"}`)
}
