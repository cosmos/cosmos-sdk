package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgUnrevokeGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress("abcd")
	msg := NewMsgUnrevoke(addr)
	bytes := msg.GetSignBytes()
	require.Equal(t, string(bytes), `{"address":"cosmosvaladdr1v93xxeqamr0mv"}`)
}
