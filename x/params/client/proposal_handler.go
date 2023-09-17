package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"cosmossdk.io/x/params/client/cli"
)

// ProposalHandler is the param change proposal handler.
var ProposalHandler = govclient.NewProposalHandler(cli.NewSubmitParamChangeProposalTxCmd)
