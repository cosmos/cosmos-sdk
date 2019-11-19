package client

import (
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
