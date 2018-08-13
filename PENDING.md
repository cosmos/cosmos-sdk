## v0.24.0 PENDING 

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  - [x/stake] \#1880 More REST-ful endpoints
  - [x/slashing] \#1866 `/slashing/signing_info` takes cosmosvalpub instead of cosmosvaladdr
  - use time.Time instead of int64 for time. 

* Gaia CLI  (`gaiacli`)
  -  [x/stake] change `--keybase-sig` to `--identity`
  -  [x/gov] Change `proposalID` to `proposal-id` 
  -  [x/stake, x/gov] \#1606 Use `--from` instead of adhoc flags like `--address-validator` 
        and `--proposer` to indicate the sender address.
  -  \#1551 Remove `--name` completely
  -  Genesis/key creation (`gaiad init`) now supports user-provided key passwords

* Gaia
  - [x/stake] Inflation doesn't use rationals in calculation (performance boost)
  - [x/stake] Persist a map from `addr->pubkey` in the state since BeginBlock
    doesn't provide pubkeys.
  - [x/gov] Added tags sub-package, changed tags to use dash-case 
  - [x/gov] Governance parameters are now stored in globalparams store
  
* SDK 
  - [baseapp] Msgs are no longer run on CheckTx, removed `ctx.IsCheckTx()`
  - [baseapp] NewBaseApp constructor now takes sdk.TxDecoder as argument instead of wire.Codec
  - [types] sdk.NewCoin now takes sdk.Int, sdk.NewInt64Coin takes int64
  - [x/auth] Default TxDecoder can be found in `x/auth` rather than baseapp

* Tendermint 
    - v0.22.5 -> [Tendermint PR](https://github.com/tendermint/tendermint/pull/1966)
        - change all the cryptography imports. 
    - v0.23.0 -> [SDK PR](https://github.com/cosmos/cosmos-sdk/pull/1927)
        - BeginBlock no longer includes crypto.Pubkey
        - use time.Time instead of int64 for time. 

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
* [networks] Added ansible scripts to upgrade seed nodes on a network
* [store] Add transient store
* [gov] Add slashing for validators who do not vote on a proposal
* [cli] added `gov query-proposals` command to CLI. Can filter by `depositer`, `voter`, and `status`
* [core] added BaseApp.Seal - ability to seal baseapp parameters once they've been set
* [scripts] added log output monitoring to DataDog using Ansible scripts
* [gov] added TallyResult type that gets added stored in Proposal after tallying is finished

IMPROVEMENTS
* [baseapp] Allow any alphanumeric character in route
* [tools] Remove `rm -rf vendor/` from `make get_vendor_deps`
* [x/auth] Recover ErrorOutOfGas panic in order to set sdk.Result attributes correctly
* [spec] \#967 Inflation and distribution specs drastically improved
* [tests] Add tests to example apps in docs
* [x/gov] Votes on a proposal can now be queried
* [x/bank] Unit tests are now table-driven
* [tests] Fixes ansible scripts to work with AWS too
* [tests] \#1806 CLI tests are now behind the build flag 'cli_test', so go test works on a new repo
* [x/gov] Initial governance parameters can now be set in the genesis file
* [x/stake] \#1815 Sped up the processing of `EditValidator` txs. 
* [server] \#1930 Transactions indexer indexes all tags by default.

BUG FIXES
*  \#1766 Fixes bad example for keybase identity
*  \#1804 Fixes gen-tx genesis generation logic temporarily until upstream updates
*  \#1799 Fix `gaiad export`
*  \#1828 Force user to specify amount on create-validator command by removing default
*  \#1839 Fixed bug where intra-tx counter wasn't set correctly for genesis validators
* [staking] [#1858](https://github.com/cosmos/cosmos-sdk/pull/1858) Fixed bug where the cliff validator was not be updated correctly
* [tests] \#1675 Fix non-deterministic `test_cover` 
* [client] \#1551: Refactored `CoreContext`
  * Renamed `CoreContext` to `QueryContext`
  * Removed all tx related fields and logic (building & signing) to separate
  structure `TxContext` in `x/auth/client/context`
  * Cleaned up documentation and API of what used to be `CoreContext`
  * Implemented `KeyType` enum for key info
* [tests] \#1551: Fixed invalid LCD test JSON payload in `doIBCTransfer`
* [basecoin] Fixes coin transaction failure and account query [discussion](https://forum.cosmos.network/t/unmarshalbinarybare-expected-to-read-prefix-bytes-75fbfab8-since-it-is-registered-concrete-but-got-0a141dfa/664/6)
