//go:build norace
// +build norace

package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	client "github.com/cosmos/cosmos-sdk/x/sanction/client/cli"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	cfg.NumValidators = 5
	cfg.TimeoutCommit = 1 * time.Second

	// Define some stuff in the sanction genesis state.
	sanctionedAddr1 := sdk.AccAddress("1_sanctioned_address_")
	sanctionedAddr2 := sdk.AccAddress("2_sanctioned_address_")
	tempSanctAddr := sdk.AccAddress("temp_sanctioned_addr")
	tempUnsanctAddr := sdk.AccAddress("temp_unsanctioned___")
	sanctionGenBz := cfg.GenesisState[sanction.ModuleName]
	var sanctionGen sanction.GenesisState
	if len(sanctionGenBz) > 0 {
		cfg.Codec.MustUnmarshalJSON(sanctionGenBz, &sanctionGen)
	}
	sanctionGen.SanctionedAddresses = append(sanctionGen.SanctionedAddresses,
		sanctionedAddr1.String(),
		sanctionedAddr2.String(),
	)
	sanctionGen.TemporaryEntries = append(sanctionGen.TemporaryEntries,
		&sanction.TemporaryEntry{
			Address:    tempSanctAddr.String(),
			ProposalId: 1,
			Status:     sanction.TEMP_STATUS_SANCTIONED,
		},
		&sanction.TemporaryEntry{
			Address:    tempUnsanctAddr.String(),
			ProposalId: 1,
			Status:     sanction.TEMP_STATUS_UNSANCTIONED,
		},
	)
	sanctionGen.Params = &sanction.Params{
		ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin(cfg.BondDenom, 52)),
		ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin(cfg.BondDenom, 133)),
	}
	cfg.GenesisState[sanction.ModuleName] = cfg.Codec.MustMarshalJSON(&sanctionGen)

	// Tweak the gov params too to make testing gov props easier.
	// MinDeposit: 6stake (default is 10000000stake)
	// MaxDepositPeriod: 5s (default is 48h)
	// VotingPeriod: 5s (default is 48h)
	govGenBz := cfg.GenesisState[gov.ModuleName]
	var govGen govv1.GenesisState
	if len(govGenBz) > 0 {
		cfg.Codec.MustUnmarshalJSON(govGenBz, &govGen)
	}
	govGen.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewInt64Coin(cfg.BondDenom, 6))
	fiveSeconds := time.Second * 5
	govGen.DepositParams.MaxDepositPeriod = &fiveSeconds
	govGen.VotingParams.VotingPeriod = &fiveSeconds
	cfg.GenesisState[gov.ModuleName] = cfg.Codec.MustMarshalJSON(&govGen)

	suite.Run(t, NewIntegrationTestSuite(cfg, &sanctionGen))
}

func (s *IntegrationTestSuite) TestSanctionValidatorImmediateUsingGovCmds() {
	// Wait 2 blocks to start this. That way, hopefully the query tests are done.
	// In between the two, create all the stuff to send.
	s.Require().NoError(s.network.WaitForNextBlock(), "wait for next block 1")
	authority := s.getAuthority()
	proposerValI := 0
	sanctionValI := 4
	sanctMsg := &sanction.MsgSanction{
		Addresses: []string{s.network.Validators[sanctionValI].Address.String()},
		Authority: authority,
	}
	sanctMsgAny, err := codectypes.NewAnyWithValue(sanctMsg)
	depAmt := s.sanctionGenesis.Params.ImmediateSanctionMinDeposit
	feeAmt := sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)
	s.Require().NoError(err, "NewAnyWithValue(MsgSanction)")
	// Thankfully, the struct used to unmarshal the proposal json (in NewCmdSubmitProposal), is private.
	// And to be really helpful, it's not the same as MsgSubmitProposal.
	// Specifically, the command wants keys "messages", "metadata", and "deposit",
	// but MsgSubmitProposal ends up with "messages", "metadata", "initial_deposit", and "proposer".
	// So I'm going to marshal MsgSubmitProposal to json, then unmarshal it to a map[string]json.RawMessage,
	// tweak the map, then marshal that into what ends up in the file.
	propMsg := &govv1.MsgSubmitProposal{
		Messages:       []*codectypes.Any{sanctMsgAny},
		InitialDeposit: depAmt,
		Proposer:       s.network.Validators[proposerValI].Address.String(),
		Metadata:       "",
	}
	propMsgBzStep1, err := s.cfg.Codec.MarshalJSON(propMsg)
	s.Require().NoError(err, "MarshalJSON(MsgSubmitProposal)")
	var propMsgJSON map[string]json.RawMessage
	err = json.Unmarshal(propMsgBzStep1, &propMsgJSON)
	s.Require().NoError(err, "Unmarshal MsgSubmitProposal")
	propMsgJSON["deposit"] = []byte(fmt.Sprintf("%q", depAmt.String()))
	delete(propMsgJSON, "initial_deposit")
	delete(propMsgJSON, "proposer")
	propMsgBz, err := json.Marshal(propMsgJSON)
	s.Require().NoError(err, "Marshal propMsgJSON")
	propFile := filepath.Join(s.T().TempDir(), "gov-prop-sanction.json")
	err = os.WriteFile(propFile, propMsgBz, 0644)
	s.Require().NoError(err, "WriteFile %s", propFile)

	// Usage: simd tx gov submit-proposal [path/to/proposal.json] [flags]
	propCmd := govcli.NewCmdSubmitProposal()
	propArgs := []string{
		propFile,
		"--" + flags.FlagKeyringBackend, keyring.BackendTest,
		"--" + flags.FlagFrom, propMsg.Proposer,
		"--" + flags.FlagFees, feeAmt.String(),
		"--" + flags.FlagSkipConfirmation,
		"--" + flags.FlagBroadcastMode, flags.BroadcastBlock,
		"--" + tmcli.OutputFlag, "json",
	}

	// Usage: simd query gov proposals [flags]
	propsQueryCmd := govcli.GetCmdQueryProposals()
	propsQueryArgs := []string{
		"--" + flags.FlagReverse,
		"--" + flags.FlagLimit, "1",
		"--" + tmcli.OutputFlag, "json",
	}

	// Usage: simd tx gov vote [proposal-id] [option] [flags]
	voteCmd := govcli.NewCmdVote()
	allVoteArgs := make([][]string, len(s.network.Validators))
	for i, val := range s.network.Validators {
		// Note: The no_with_veto vote from the validator being sanctioned should fail because
		// enough deposit is provided to make the sanction immediate, so they won't be able to pay fees.
		// The command won't return an error though. That failure will happen when the block is being processed.
		// Failure of that tx will be reflected in the final tally of the proposal later on though,
		// i.e. it won't have any recorded no-with-veto votes.
		option := "yes"
		if i == sanctionValI {
			option = "no_with_veto"
		}
		// Note: arg[0] will be updated with the gov prop once it's known.
		allVoteArgs[i] = []string{
			"0", option,
			"--" + flags.FlagKeyringDir, filepath.Join(val.Dir, "simcli"),
			"--" + flags.FlagKeyringBackend, keyring.BackendTest,
			"--" + flags.FlagFrom, val.Address.String(),
			"--" + flags.FlagFees, feeAmt.String(),
			"--" + flags.FlagBroadcastMode, flags.BroadcastAsync,
			"--" + flags.FlagSkipConfirmation,
			"--" + tmcli.OutputFlag, "json",
		}
	}

	// Usage: simd query gov proposal [proposal-id] [flags]
	propQueryCmd := govcli.GetCmdQueryProposal()
	// Here too, that first arg will be updated when we know the proposal id.
	propQueryArgs := []string{
		"0",
		"--" + tmcli.OutputFlag, "json",
	}

	// Usage: simd query sanction is-sanctioned <address> [flags]
	isSanctCmd := client.QueryIsSanctionedCmd()
	isSanctArgs := []string{
		s.network.Validators[sanctionValI].Address.String(),
		"--" + tmcli.OutputFlag, "json",
	}

	// Finally, wait for the next block.
	s.Require().NoError(s.network.WaitForNextBlock(), "wait for next block 2")
	s.logHeight()

	// Submit the proposal. This shouldn't return until the next block is cut.
	s.T().Logf("Proposal: %s\n%s", propFile, propMsgBz)
	propOutBW, err := cli.ExecTestCLICmd(s.clientCtx, propCmd, propArgs)
	s.Require().NoError(err, "ExecTestCLICmd tx gov submit-proposal")
	propOutBz := propOutBW.Bytes()
	s.T().Logf("tx gov submit-proposal output:\n%s", propOutBz)
	propHeight := s.logHeight()

	// Find the last proposal (assuming it's the one just submitted above).
	propsQueryOutBW, err := cli.ExecTestCLICmd(s.clientCtx, propsQueryCmd, propsQueryArgs)
	s.Require().NoError(err, "ExecTestCLICmd query gov proposals")
	propsQueryOutBz := propsQueryOutBW.Bytes()
	s.T().Logf("q gov proposals output:\n%s", propsQueryOutBz)
	var propsQueryOut govv1.QueryProposalsResponse
	err = s.cfg.Codec.UnmarshalJSON(propsQueryOutBz, &propsQueryOut)
	s.Require().NoError(err, "Unmarshal QueryProposalsResponse")
	s.Require().NotEmpty(propsQueryOut.Proposals, "proposals")
	s.Assert().Equal(govv1.StatusVotingPeriod, propsQueryOut.Proposals[0].Status, "proposal status")
	propID := fmt.Sprintf("%d", propsQueryOut.Proposals[0].Id)
	s.T().Logf("Proposal id to vote on: %s", propID)

	// Verify that the validator is sanctioned
	isSanctOutBW1, err := cli.ExecTestCLICmd(s.clientCtx, isSanctCmd, isSanctArgs)
	s.Require().NoError(err, "ExecTestCLICmd query sanction is-sanctioned (first time)")
	isSanctOutBz1 := isSanctOutBW1.Bytes()
	s.T().Logf("query sanction is-sanctioned output (first time):\n%s", isSanctOutBz1)
	var isSanctOut1 sanction.QueryIsSanctionedResponse
	err = json.Unmarshal(isSanctOutBz1, &isSanctOut1)
	s.Require().NoError(err, "Unmarshal QueryIsSanctionedResponse (first time)")
	s.Assert().True(isSanctOut1.IsSanctioned, "is sanctioned (first time)")

	// Cast votes on it.
	for i, voteArgs := range allVoteArgs {
		voteArgs[0] = propID
		voteOutBW, err := cli.ExecTestCLICmd(s.clientCtx, voteCmd, voteArgs)
		s.Require().NoError(err, "[%d]: ExecTestCLICmd tx gov vote", i)
		voteOutBz := voteOutBW.Bytes()
		s.T().Logf("[%d]: tx gov vote output:\n%s", i, voteOutBz)
	}

	// We configured 1 second per block, and a 5-second voting period.
	// So wait for 6 blocks after the proposal block.
	s.logHeight()
	s.T().Log("waiting for voting period to end")
	_, err = s.network.WaitForHeight(propHeight + 6)
	s.Require().NoError(err, "waiting for block after voting should end")
	lastHeight := s.logHeight()

	// Check that the proposal passed.
	propQueryArgs[0] = propID
	propQueryOutBW, err := cli.ExecTestCLICmd(s.clientCtx, propQueryCmd, propQueryArgs)
	s.Require().NoError(err, "ExecTestCLICmd query gov proposal %s", propID)
	propQueryOutBz := propQueryOutBW.Bytes()
	s.T().Logf("query gov prop %s output:\n%s", propID, propQueryOutBz)
	var propQueryOut govv1.Proposal
	err = s.cfg.Codec.UnmarshalJSON(propQueryOutBz, &propQueryOut)
	s.Require().NoError(err, "Unmarshal QueryProposalResponse")
	s.Assert().Equal(govv1.StatusPassed, propQueryOut.Status, "proposal status")

	// Check that that validator is still sanctioned.
	isSanctOutBW2, err := cli.ExecTestCLICmd(s.clientCtx, isSanctCmd, isSanctArgs)
	s.Require().NoError(err, "ExecTestCLICmd query sanction is-sanctioned (second time)")
	isSanctOutBz2 := isSanctOutBW2.Bytes()
	s.T().Logf("query sanction is-sanctioned output (second time):\n%s", isSanctOutBz2)
	var isSanctOut2 sanction.QueryIsSanctionedResponse
	err = json.Unmarshal(isSanctOutBz2, &isSanctOut2)
	s.Require().NoError(err, "Unmarshal QueryIsSanctionedResponse (second time)")
	s.Assert().True(isSanctOut2.IsSanctioned, "is sanctioned (second time)")

	// Wait 20 more blocks to make sure nothing unravels.
	s.logHeight()
	s.T().Log("waiting 20 blocks before shutdown")
	_, err = s.network.WaitForHeightWithTimeout(lastHeight+20, 30*time.Second)
	s.Require().NoError(err, "waiting for block %d (or 30 seconds)", lastHeight+20)
	s.logHeight()

	// Check that that validator is still sanctioned one last time.
	isSanctOutBW3, err := cli.ExecTestCLICmd(s.clientCtx, isSanctCmd, isSanctArgs)
	s.Require().NoError(err, "ExecTestCLICmd query sanction is-sanctioned (third time)")
	isSanctOutBz3 := isSanctOutBW3.Bytes()
	s.T().Logf("query sanction is-sanctioned output (third time):\n%s", isSanctOutBz3)
	var isSanctOut3 sanction.QueryIsSanctionedResponse
	err = json.Unmarshal(isSanctOutBz3, &isSanctOut3)
	s.Require().NoError(err, "Unmarshal QueryIsSanctionedResponse (third time)")
	s.Assert().True(isSanctOut3.IsSanctioned, "is sanctioned (third time)")

	s.T().Log("done")
}

func (s *IntegrationTestSuite) logHeight() int64 {
	height, err := s.network.LatestHeight()
	s.Require().NoError(err, "LatestHeight()")
	s.T().Logf("Current height: %d", height)
	return height
}
