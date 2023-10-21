package client

import (
	"cosmossdk.io/x/params/client/cli"

	govclient "cosmossdk.io/x/gov/client"
)

// ProposalHandler is the param change proposal handler.
var ProposalHandler = govclient.NewProposalHandler(cli.NewSubmitParamChangeProposalTxCmd)
