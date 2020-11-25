package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"

	govutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
)

func parseSubmitProposalFlags(fs *pflag.FlagSet) (*proposal, error) {
	proposal := &proposal{}
	proposalFile, _ := fs.GetString(FlagProposal)

	if proposalFile == "" {
		proposalType, _ := fs.GetString(FlagProposalType)

		proposal.Title, _ = fs.GetString(FlagTitle)
		proposal.Description, _ = fs.GetString(FlagDescription)
		proposal.Type = govutils.NormalizeProposalType(proposalType)
		proposal.Deposit, _ = fs.GetString(FlagDeposit)
		return proposal, nil
	}

	for _, flag := range ProposalFlags {
		if v, _ := fs.GetString(flag); v != "" {
			return nil, fmt.Errorf("--%s flag provided alongside --proposal, which is a noop", flag)
		}
	}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, proposal)
	if err != nil {
		return nil, err
	}

	return proposal, nil
}
