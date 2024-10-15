package systemtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	validMetadata = "metadata"
)

func TestGroupCommands(t *testing.T) {
	// scenario: test group commands
	// given a running chain

	sut.ResetChain(t)
	require.GreaterOrEqual(t, sut.NodesCount(), 2)

	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	sut.StartChain(t)

	// test create group
	memberWeight := "3"
	validMembers := fmt.Sprintf(`
	{
		"members": [
			{
				"address": "%s",
				"weight": "%s",
				"metadata": "%s"
			}
		]
	}`, valAddr, memberWeight, validMetadata)
	validMembersFile := StoreTempFile(t, []byte(validMembers))
	createGroupCmd := []string{"tx", "group", "create-group", valAddr, validMetadata, validMembersFile.Name(), "--from=" + valAddr}
	rsp := cli.RunAndWait(createGroupCmd...)
	RequireTxSuccess(t, rsp)

	// query groups by admin to confirm group creation
	rsp = cli.CustomQuery("q", "group", "groups-by-admin", valAddr)
	require.Len(t, gjson.Get(rsp, "groups").Array(), 1)
	// groupID := gjson.Get(rsp, "groups.0.id").String()

	// test create group policies

	for i := 0; i < 5; i++ {
		threshold := i + 1
		if threshold > 3 {
			threshold = 3
		}

	}
}
