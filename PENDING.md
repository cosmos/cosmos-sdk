## PENDING

BREAKING CHANGES
* [baseapp] Msgs are no longer run on CheckTx, removed `ctx.IsCheckTx()`
* [x/gov] CLI flag changed from `proposalID` to `proposal-id`
* [x/stake] Fixed the period check for the inflation calculation
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
* Added support for cosmos-sdk-cli tool under cosmos-sdk/cmd
   * This allows SDK users to init a new project repository with a single command.
* Optionally run tests within a docker container.

IMPROVEMENTS
* [baseapp] Allow any alphanumeric character in route
* [cli] Improve error messages for all txs when the account doesn't exist
* [tools] Remove `rm -rf vendor/` from `make get_vendor_deps`
* [x/stake] Add revoked to human-readable validator
* [x/gov] Votes on a proposal can now be queried
* [x/bank] Unit tests are now table-driven
* [x/slashing] No longer wait until SignedBlocksWindow has passed before downtime slashing commences.

BUG FIXES
*  \#1666 Add intra-tx counter to the genesis validators
