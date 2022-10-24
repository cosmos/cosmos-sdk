package client

import (
	"github.com/pointnetwork/cosmos-point-sdk/x/distribution/client/cli"
	govclient "github.com/pointnetwork/cosmos-point-sdk/x/gov/client"
)

// ProposalHandler is the community spend proposal handler.
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal)
)
