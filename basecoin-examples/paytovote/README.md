# paytovote plugin

### Description
paytovote is a basic application which demonstrates how to leverage the basecoin library to create an instance of the basecoin system which utilizes a custom paytovote plugin. The premise of this plugin is to allow users to pay a fee to create or vote for user-specified issues. Unique fees are applied when voting or creating a new issue. Fees may use coin types (for example "voteTokens" or "newIssueTokens"). Currently, the fee to cast a vote is decided by the user when the issue is being generated, and the fee to create a new issue is defined globally within the plugin CLI commands (cmd/paytovote/commands)

### Usage
 - enable the paytovote plugin and start `paytovote start --paytovote-plugin` 
 - start tendermint in another terminal `tendermint node`
 - create issues with `paytovote AppTx P2VCreateIssue`
   - for a complete list of required flags and usage see:
     - `paytovote AppTx -h`
     - `paytovote AppTx P2VCreateIssue -h`
 - vote for issues with `paytovote AppTx P2VVote`
   - for a complete list of required flags and usage see:
     - `paytovote AppTx -h`
     - `paytovote AppTx P2VVote -h`
