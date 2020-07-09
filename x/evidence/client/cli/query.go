package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

// GetQueryCmd returns the CLI command with all evidence module query commands
// mounted.
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Query for evidence by hash or for all (paginated) submitted evidence",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for specific submitted evidence by hash or query for all (paginated) evidence:
	
Example:
$ %s query %s DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660
$ %s query %s --page=2 --limit=50
`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		Args:                       cobra.MaximumNArgs(1),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       QueryEvidenceCmd(cdc),
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of evidence to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of evidence to query for")

	return flags.GetCommands(cmd)[0]
}

// QueryEvidenceCmd returns the command handler for evidence querying. Evidence
// can be queried for by hash or paginated evidence can be returned.
func QueryEvidenceCmd(cdc *codec.Codec) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := client.ValidateCmd(cmd, args); err != nil {
			return err
		}

		clientCtx := client.NewContext().WithCodec(cdc).WithJSONMarshaler(cdc)

		if hash := args[0]; hash != "" {
			return queryEvidence(cdc, clientCtx, hash)
		}

		page, _ := cmd.Flags().GetInt(flags.FlagPage)
		limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

		return queryAllEvidence(clientCtx, page, limit)
	}
}

func queryEvidence(cdc *codec.Codec, clientCtx client.Context, hash string) error {
	if _, err := hex.DecodeString(hash); err != nil {
		return fmt.Errorf("invalid evidence hash: %w", err)
	}

	params := types.NewQueryEvidenceParams(hash)
	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryEvidence)
	res, _, err := clientCtx.QueryWithData(route, bz)
	if err != nil {
		return err
	}

	var evidence exported.Evidence
	err = cdc.UnmarshalJSON(res, &evidence)
	if err != nil {
		return fmt.Errorf("failed to unmarshal evidence: %w", err)
	}

	return clientCtx.PrintOutput(evidence)
}

func queryAllEvidence(clientCtx client.Context, page, limit int) error {
	params := types.NewQueryAllEvidenceParams(page, limit)
	bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
	if err != nil {
		return fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllEvidence)
	res, _, err := clientCtx.QueryWithData(route, bz)
	if err != nil {
		return err
	}

	var evidence []exported.Evidence
	err = clientCtx.JSONMarshaler.UnmarshalJSON(res, &evidence)
	if err != nil {
		return fmt.Errorf("failed to unmarshal evidence: %w", err)
	}

	return clientCtx.PrintOutput(evidence)
}
