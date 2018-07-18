## PENDING

BREAKING CHANGES
* [baseapp] Msgs are no longer run on CheckTx, removed `ctx.IsCheckTx()`
* [x/gov] CLI flag changed from `proposalID` to `proposal-id`
* [x/stake] Inflation doesn't use rationals in calculation (performance boost)
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
* [x/gov] Added tags sub-package, changed tags to use dash-case 

FEATURES
* [lcd] Can now query governance proposals by ProposalStatus
* [x/mock/simulation] Randomized simulation framework
  * Modules specify invariants and operations, preferably in an x/[module]/simulation package
  * Modules can test random combinations of their own operations
  * Applications can integrate operations and invariants from modules together for an integrated simulation
* [baseapp] Initialize validator set on ResponseInitChain
* [cosmos-sdk-cli] Added support for cosmos-sdk-cli tool under cosmos-sdk/cmd	
   * This allows SDK users to initialize a new project repository.
* [tests] Remotenet commands for AWS (awsnet)

IMPROVEMENTS
* [baseapp] Allow any alphanumeric character in route
* [tools] Remove `rm -rf vendor/` from `make get_vendor_deps`
* [x/auth] Recover ErrorOutOfGas panic in order to set sdk.Result attributes correctly
* [tests] Add tests to example apps in docs
* [x/gov] Votes on a proposal can now be queried
* [x/bank] Unit tests are now table-driven
* [tests] Fixes ansible scripts to work with AWS too
