package client

import (
	govclient "github.com/Stride-Labs/cosmos-sdk/x/gov/client"
	"github.com/Stride-Labs/cosmos-sdk/x/params/client/cli"
	"github.com/Stride-Labs/cosmos-sdk/x/params/client/rest"
)

// ProposalHandler is the param change proposal handler.
var ProposalHandler = govclient.NewProposalHandler(cli.NewSubmitParamChangeProposalTxCmd, rest.ProposalRESTHandler)
