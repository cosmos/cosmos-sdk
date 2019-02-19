package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
)

func TestResult(t *testing.T) {
	var res Result
	require.True(t, res.IsOK())

	res.Data = []byte("data")
	require.True(t, res.IsOK())

	res.Code = CodeType(1)
	require.False(t, res.IsOK())
}

func TestTxResponseJSON(t *testing.T) {
	cdc := codec.New()
	txr := TxResponse{
		Log: `[{"log":"","msg_index":"0","success":true}]`,
	}

	bz, err := cdc.MarshalJSON(txr)
	require.NoError(t, err)

	var txr2 TxResponse
	err = cdc.UnmarshalJSON(bz, &txr2)
	require.NoError(t, err)

	require.Equal(t, txr, txr2)
}
