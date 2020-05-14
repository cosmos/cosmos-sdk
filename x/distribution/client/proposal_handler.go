package client

import (
	"github.com/KiraCore/cosmos-sdk/x/distribution/client/cli"
	"github.com/KiraCore/cosmos-sdk/x/distribution/client/rest"
	govclient "github.com/KiraCore/cosmos-sdk/x/gov/client"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
