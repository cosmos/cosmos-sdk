package cli

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/x/params"
)

type proposal struct {
	Title       string
	Description string
	Deposit     string
	Changes     []params.Change
}

const flagProposal = "proposal"

// Only works for external JSON file
func parseSubmitProposalJSON() (*proposal, error) {
	proposal := &proposal{}

	proposalFile := viper.GetString(flagProposal)

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
