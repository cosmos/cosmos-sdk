package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/evidence"
)

const (
	flagPage  = "page"
	flagLimit = "limit"
)

// GetQueryCmd returns the CLI command with all evidence module query commands
// mounted.
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   evidence.ModuleName,
		Short: "Query for evidence by hash or for all (paginated) submitted evidence",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for specific submitted evidence by hash or query for all (paginated) evidence:
	
Example:
$ %s query %s DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660
$ %s query %s --page=2 --limit=50
`,
				version.ClientName, evidence.ModuleName, version.ClientName, evidence.ModuleName,
			),
		),
		Args:                       cobra.MaximumNArgs(1),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       QueryEvidenceCMD(cdc),
	}

	cmd.Flags().Int(flagPage, 1, "pagination page of evidence to to query for")
	cmd.Flags().Int(flagLimit, 100, "pagination limit of evidence to query for")

	return cmd
}

// QueryEvidenceCMD returns the command handler for evidence querying. Evidence
// can be queried for by hash or paginated evidence can be returned.
func QueryEvidenceCMD(cdc *codec.Codec) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := client.ValidateCmd(cmd, args); err != nil {
			return err
		}

		cliCtx := context.NewCLIContext().WithCodec(cdc)

		if hash := args[0]; hash != "" {
			return queryEvidence(cdc, cliCtx, hash)
		}

		return queryAllEvidence(cdc, cliCtx)
	}
}

func queryEvidence(cdc *codec.Codec, cliCtx context.CLIContext, hash string) error {
	if _, err := hex.DecodeString(hash); err != nil {
		return fmt.Errorf("invalid evidence hash: %w", err)
	}

	params := evidence.NewQueryEvidenceParams(hash)

	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s", evidence.QuerierRoute, evidence.QueryEvidence)
	res, _, err := cliCtx.QueryWithData(route, bz)
	if err != nil {
		return err
	}

	var evidence evidence.Evidence
	err = cdc.UnmarshalJSON(res, &evidence)
	if err != nil {
		return fmt.Errorf("failed to unmarshal evidence: %w", err)
	}

	return cliCtx.PrintOutput(evidence)
}

func queryAllEvidence(cdc *codec.Codec, cliCtx context.CLIContext) error {
	params := evidence.NewQueryAllEvidenceParams(viper.GetInt(flagPage), viper.GetInt(flagLimit))

	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s", evidence.QuerierRoute, evidence.QueryAllEvidence)
	res, _, err := cliCtx.QueryWithData(route, bz)
	if err != nil {
		return err
	}

	var evidence []evidence.Evidence
	err = cdc.UnmarshalJSON(res, &evidence)
	if err != nil {
		return fmt.Errorf("failed to unmarshal evidence: %w", err)
	}

	return cliCtx.PrintOutput(evidence)
}
