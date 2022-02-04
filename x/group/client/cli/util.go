package cli

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// CLIProposal defines a Msg-based group proposal for CLI purposes.
type CLIProposal struct {
	GroupPolicyAddress string
	// Messages defines an array of sdk.Msgs proto-JSON-encoded as Anys.
	Messages  []json.RawMessage
	Metadata  []byte
	Proposers []string
}

func parseCLIProposal(path string) (CLIProposal, error) {
	var p CLIProposal

	contents, err := os.ReadFile(path)
	if err != nil {
		return CLIProposal{}, err
	}

	err = json.Unmarshal(contents, &p)
	if err != nil {
		return CLIProposal{}, err
	}

	return p, nil
}

func parseMsgs(cdc codec.Codec, p CLIProposal) ([]sdk.Msg, error) {
	msgs := make([]sdk.Msg, len(p.Messages))
	for i, anyJSON := range p.Messages {
		var msg sdk.Msg
		err := cdc.UnmarshalInterfaceJSON(anyJSON, &msg)
		if err != nil {
			return nil, err
		}

		msgs[i] = msg
	}

	return msgs, nil
}
