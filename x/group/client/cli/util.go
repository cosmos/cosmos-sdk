package cli

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/x/group"
)

func parseMembers(clientCtx client.Context, membersFile string) ([]group.Member, error) {
	members := group.Members{}

	if membersFile == "" {
		return members.Members, nil
	}

	contents, err := ioutil.ReadFile(membersFile)
	if err != nil {
		return nil, err
	}

	err = clientCtx.Codec.UnmarshalJSON(contents, &members)
	if err != nil {
		return nil, err
	}

	return members.Members, nil
}

func execFromString(execStr string) group.Exec {
	exec := group.Exec_EXEC_UNSPECIFIED
	switch execStr {
	case ExecTry:
		exec = group.Exec_EXEC_TRY
	}
	return exec
}
