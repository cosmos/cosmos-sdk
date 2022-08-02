package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

func parseDecisionPolicy(cdc codec.Codec, decisionPolicyFile string) (group.DecisionPolicy, error) {
	if decisionPolicyFile == "" {
		return nil, fmt.Errorf("decision policy is required")
	}

	contents, err := ioutil.ReadFile(decisionPolicyFile)
	if err != nil {
		return nil, err
	}

	var policy group.DecisionPolicy
	if err := cdc.UnmarshalInterfaceJSON(contents, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse decision policy: %w", err)
	}

	return policy, nil
}

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
type CLIProposal struct {
	GroupPolicyAddress string `json:"group_policy_address"`
	// Messages defines an array of sdk.Msgs proto-JSON-encoded as Anys.
	Messages  []json.RawMessage `json:"messages"`
	Metadata  string            `json:"metadata"`
	Proposers []string          `json:"proposers"`
}

func getCLIProposal(path string) (CLIProposal, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return CLIProposal{}, err
	}

	return parseCLIProposal(contents)
}

func parseCLIProposal(contents []byte) (CLIProposal, error) {
	var p CLIProposal
	if err := json.Unmarshal(contents, &p); err != nil {
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
