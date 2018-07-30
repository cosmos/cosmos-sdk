## PENDING

BREAKING CHANGES
* Update to tendermint v0.22.5. This involves changing all of the cryptography imports. [Ref](https://github.com/tendermint/tendermint/pull/1966)
* [baseapp] Msgs are no longer run on CheckTx, removed `ctx.IsCheckTx()`
* [x/gov] CLI flag changed from `proposalID` to `proposal-id`
* [x/stake] Fixed the period check for the inflation calculation
* [x/stake] Inflation doesn't use rationals in calculation (performance boost)
* [x/stake] CLI flags for identity changed from `--keybase-sig` to `--identity`, effects:
  * `gaiacli stake create-validator`
  * `gaiacli stake edit-validator`
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
* [x/gov] Governance parameters are now stored in globalparams store

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
* [store] Add transient store

IMPROVEMENTS
* [baseapp] Allow any alphanumeric character in route
* [cli] Improve error messages for all txs when the account doesn't exist
* [tools] Remove `rm -rf vendor/` from `make get_vendor_deps`
* [x/auth] Recover ErrorOutOfGas panic in order to set sdk.Result attributes correctly
* [x/stake] Add revoked to human-readable validator 
* [tests] Add tests to example apps in docs
* [x/gov] Votes on a proposal can now be queried
* [x/bank] Unit tests are now table-driven
* [tests] Fixes ansible scripts to work with AWS too
* [tests] \#1806 CLI tests are now behind the build flag 'cli_test', so go test works on a new repo
* [x/gov] Initial governance parameters can now be set in the genesis file
* [x/stake] \#1815 Sped up the processing of `EditValidator` txs. 

BUG FIXES
*  \#1666 Add intra-tx counter to the genesis validators
*  \#1797 Fix off-by-one error in slashing for downtime
*  \#1787 Fixed bug where Tally fails due to revoked/unbonding validator
*  \#1766 Fixes bad example for keybase identity
*  \#1804 Fixes gen-tx genesis generation logic temporarily until upstream updates
*  \#1799 Fix `gaiad export`
*  \#1828 Force user to specify amount on create-validator command by removing default
*  \#1839 Fixed bug where intra-tx counter wasn't set correctly for genesis validators
* [tests] \#1675 Fix non-deterministic `test_cover` 
