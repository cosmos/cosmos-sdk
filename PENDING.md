## PENDING

BREAKING CHANGES
* [baseapp] Msgs are no longer run on CheckTx, removed `ctx.IsCheckTx()`
* [x/stake] Fixed the period check for the inflation calculation
* [baseapp] NewBaseApp constructor now takes sdk.TxDecoder as argument instead of wire.Codec
* [x/auth] Default TxDecoder can be found in `x/auth` rather than baseapp
* \#1606 The following CLI commands have been switched to use `--from`
  * `gaiacli stake create-validator --address-validator`
  * `gaiacli stake edit-validator --address-validator`
  * `gaiacli stake delegate --address-delegator`
  * `gaiacli stake unbond begin --address-delegator`
  * `gaiacli stake unbond complete --address-delegator`
  * `gaiacli stake redelegate begin --address-delegator`
  * `gaiacli stake redelegate complete --address-delegator`
  * `gaiacli stake unrevoke [validator-address]`
  * `gaiacli gov submit-proposal --proposer`
  * `gaiacli gov deposit --depositer`
  * `gaiacli gov vote --voter`

FEATURES
* [lcd] Can now query governance proposals by ProposalStatus
* [baseapp] Initialize validator set on ResponseInitChain
* Added support for cosmos-sdk-cli tool under cosmos-sdk/cmd	
   * This allows SDK users to init a new project repository with a single command.

IMPROVEMENTS
* [baseapp] Allow any alphanumeric character in route
* [cli] Improve error messages for all txs when the account doesn't exist
* [tools] Remove `rm -rf vendor/` from `make get_vendor_deps`
* [x/stake] Add revoked to human-readable validator
* [x/auth] Recover ErrorOutOfGas panic in order to set sdk.Result attributes correctly

BUG FIXES
*  \#1666 Add intra-tx counter to the genesis validators
