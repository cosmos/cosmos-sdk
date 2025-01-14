package tx

import (
	"testing"

	any "github.com/cosmos/gogoproto/types/any"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/tx"
)

func TestIntoV2SignerInfo(t *testing.T) {
	require.NotNil(t, intoV2SignerInfo([]*tx.SignerInfo{{}}))
	require.NotNil(t, intoV2SignerInfo([]*tx.SignerInfo{{PublicKey: &any.Any{}}}))
}
