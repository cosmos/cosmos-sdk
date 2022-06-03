package cli

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

func parseMembers(membersFile string) ([]group.MemberRequest, error) {
	members := group.MemberRequests{}

	if membersFile == "" {
		return members.Members, nil
	}

	contents, err := ioutil.ReadFile(membersFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &members)
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
type Proposal struct {
	GroupPolicyAddress string `json:"group_policy_address"`
	// Messages defines an array of sdk.Msgs proto-JSON-encoded as Anys.
	Messages  []json.RawMessage `json:"messages"`
	Metadata  string            `json:"metadata"`
	Proposers []string          `json:"proposers"`
}

func getCLIProposal(path string) (Proposal, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Proposal{}, err
	}

	return parseCLIProposal(contents)
}

func parseCLIProposal(contents []byte) (Proposal, error) {
	var p Proposal
	if err := json.Unmarshal(contents, &p); err != nil {
		return Proposal{}, err
	}

	return p, nil
}

func parseMsgs(cdc codec.Codec, p Proposal) ([]sdk.Msg, error) {
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
