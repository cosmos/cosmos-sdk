package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
)

func TestParseABCILog(t *testing.T) {
	logs := `[{"log":"","msg_index":1,"success":true}]`

	res, err := ParseABCILogs(logs)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, res[0].Log, "")
	require.Equal(t, res[0].MsgIndex, uint16(1))
}

func TestABCIMessageLog(t *testing.T) {
	events := Events{NewEvent("transfer", NewAttribute("sender", "foo"))}
	msgLog := NewABCIMessageLog(0, "", events)

	msgLogs := ABCIMessageLogs{msgLog}
	bz, err := codec.Cdc.MarshalJSON(msgLogs)
	require.NoError(t, err)
	require.Equal(t, string(bz), msgLogs.String())
}
