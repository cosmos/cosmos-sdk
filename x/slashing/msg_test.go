package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgUnjailGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress("abcd")
	msg := NewMsgUnjail(addr)
	bytes := msg.GetSignBytes()
	require.Equal(t, string(bytes), `{"address":"cosmosaccaddr1v93xxeqhyqz5v"}`)
}
