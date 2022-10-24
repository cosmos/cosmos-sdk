package client

import (
	govclient "github.com/pointnetwork/cosmos-point-sdk/x/gov/client"
	"github.com/pointnetwork/cosmos-point-sdk/x/params/client/cli"
)

// ProposalHandler is the param change proposal handler.
var ProposalHandler = govclient.NewProposalHandler(cli.NewSubmitParamChangeProposalTxCmd)
