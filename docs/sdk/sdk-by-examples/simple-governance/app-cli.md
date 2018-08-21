## Application CLI 

**File: [`cmd/simplegovcli/maing.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/cmd/simplegovcli/main.go)**

To interact with our application, let us add the commands from the `simple_governance` module to our `simpleGov` application, as well as the pre-built SDK commands:

```go
//  cmd/simplegovcli/main.go
...
	rootCmd.AddCommand(
		client.GetCommands(
			simplegovcmd.GetCmdQueryProposal("proposals", cdc),
			simplegovcmd.GetCmdQueryProposals("proposals", cdc),
			simplegovcmd.GetCmdQueryProposalVotes("proposals", cdc),
			simplegovcmd.GetCmdQueryProposalVote("proposals", cdc),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			simplegovcmd.PostCmdPropose(cdc),
			simplegovcmd.PostCmdVote(cdc),
		)...)
...
```