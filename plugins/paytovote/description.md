# paytovote plugin

### Description
paytovote is a basic application which demonstrates how to leverage the basecoin library to create an instance of the basecoin system which utilizes a custom paytovote plugin. The premise of this plugin is to allow users to pay a fee to create or vote for user-specified issues. When implementing this plugin, the fee associated with voting may separate from the fee associated for creating a new issue. Additionally, each fee may utilize custom and unique token types (for example "voteTokens" or "newIssueTokens"). 

### Use
A good way to get a general sense of the technical implementation of a paytovote system is to check out the test file which can be found under basecoin/plugins/paytovote/paytovote\_test.go. The application specific transaction data which is sent through the AppTx.Data term is as follow:
 - Valid (bool) 
   - Transactions will only run if this term is true
 - Issue (string) 
   - Name of the issue which is being voted for or created
 - ActionTypeByte (byte)
   - TypeByte field which specifies the action to be taken by the paytovote transaction
   - Available actions:
     - Create a non-existent issue
     - submit a vote for an existing issue
     - submit a vote against an existing issue
 - CostToVote (types.Coins)
   - The cost charged by the plugin to submit a vote on an existing issue
 - CostToCreateIssue (types.Coins) 
   - The cost charged by the plugin when creating a new issue

