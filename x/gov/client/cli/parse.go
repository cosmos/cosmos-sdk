package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/viper"

	govClientUtils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
)

func readProposalType() string {
	return govClientUtils.NormalizeProposalType(viper.GetString(flagProposalType))
}

func parseSubmitProposalFlags() (*proposal, error) {
	proposal := &proposal{}
	proposalFile := viper.GetString(flagProposal)

	if proposalFile == "" {
		proposal.Title = viper.GetString(flagTitle)
		proposal.Description = viper.GetString(flagDescription)
		proposal.Type = readProposalType()
		proposal.Deposit = viper.GetString(flagDeposit)

		if proposal.Type == "" {
			return nil, fmt.Errorf("Invalid ProposalType %s was given", flagProposalType)
		}
		return proposal, nil
	}

	for _, flag := range proposalFlags {
		if viper.GetString(flag) != "" {
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
