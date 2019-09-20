package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/require"
)

func TestResult(t *testing.T) {
	var res Result
	require.True(t, res.IsOK())

	res.Data = []byte("data")
	require.True(t, res.IsOK())

	res.Code = CodeType(1)
	require.False(t, res.IsOK())
}

func TestParseABCILog(t *testing.T) {
	logs := `[{"log":"","msg_index":1,"success":true}]`

	res, err := ParseABCILogs(logs)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, res[0].Log, "")
	require.Equal(t, res[0].MsgIndex, uint16(1))
	require.True(t, res[0].Success)
}

func TestABCIMessageLog(t *testing.T) {
	events := Events{NewEvent("transfer", NewAttribute("sender", "foo"))}
	msgLog := NewABCIMessageLog(0, true, "", events)

	msgLogs := ABCIMessageLogs{msgLog}
	bz, err := codec.Cdc.MarshalJSON(msgLogs)
	require.NoError(t, err)
	require.Equal(t, string(bz), msgLogs.String())
}
