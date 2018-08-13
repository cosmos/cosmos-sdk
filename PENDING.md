## v0.24.0 PENDING 

* Gaia REST API (`gaiacli advanced rest-server`)
* Gaia CLI  (`gaiacli`)
* Gaia
* SDK 
* Tendermint 

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  - [x/stake] \#1880 More REST-ful endpoints (large refactor)
  - [x/slashing] \#1866 `/slashing/signing_info` takes cosmosvalpub instead of cosmosvaladdr
  - use time.Time instead of int64 for time. 

* Gaia CLI  (`gaiacli`)
  -  [x/stake] change `--keybase-sig` to `--identity`
  -  [x/stake] \#1828 Force user to specify amount on create-validator command by removing default
  -  [x/gov] Change `--proposalID` to `--proposal-id` 
  -  [x/stake, x/gov] \#1606 Use `--from` instead of adhoc flags like `--address-validator` 
        and `--proposer` to indicate the sender address.
  -  \#1551 Remove `--name` completely
  -  Genesis/key creation (`gaiad init`) now supports user-provided key passwords

* Gaia
  - [x/stake] Inflation doesn't use rationals in calculation (performance boost)
  - [x/stake] Persist a map from `addr->pubkey` in the state since BeginBlock
    doesn't provide pubkeys.
  - [x/gov] \#1781 Added tags sub-package, changed tags to use dash-case 
  - [x/gov] \#1688 Governance parameters are now stored in globalparams store
  - [x/gov] \#1859 Slash validators who do not vote on a proposal
  - [x/gov] \#1914 added TallyResult type that gets stored in Proposal after tallying is finished
  
* SDK 
  - [baseapp] Msgs are no longer run on CheckTx, removed `ctx.IsCheckTx()`
  - [baseapp] NewBaseApp constructor takes sdk.TxDecoder as argument instead of wire.Codec
  - [types] sdk.NewCoin takes sdk.Int, sdk.NewInt64Coin takes int64
  - [x/auth] Default TxDecoder can be found in `x/auth` rather than baseapp
  - [client] \#1551: Refactored `CoreContext` to `TxContext` and `QueryContext`
      - Removed all tx related fields and logic (building & signing) to separate
        structure `TxContext` in `x/auth/client/context`

* Tendermint 
    - v0.22.5 -> [Tendermint PR](https://github.com/tendermint/tendermint/pull/1966)
        - change all the cryptography imports. 
    - v0.23.0 -> [SDK PR](https://github.com/cosmos/cosmos-sdk/pull/1927)
        - BeginBlock no longer includes crypto.Pubkey
        - use time.Time instead of int64 for time. 

FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
    - [x/gov] Can now query governance proposals by ProposalStatus

* Gaia CLI  (`gaiacli`)
    - [x/gov] added `query-proposals` command. Can filter by `depositer`, `voter`, and `status`

* Gaia
  - [networks] Added ansible scripts to upgrade seed nodes on a network

* SDK 
  - [x/mock/simulation] Randomized simulation framework
     - Modules specify invariants and operations, preferably in an x/[module]/simulation package
     - Modules can test random combinations of their own operations
     - Applications can integrate operations and invariants from modules together for an integrated simulation
  - [store] \#1481 Add transient store
  - [baseapp] Initialize validator set on ResponseInitChain
  - [baseapp] added BaseApp.Seal - ability to seal baseapp parameters once they've been set
  - [cosmos-sdk-cli] New `cosmos-sdk-cli` tool to quickly initialize a new
    SDK-based project
  - [scripts] added log output monitoring to DataDog using Ansible scripts

IMPROVEMENTS

* Gaia
  - [spec] \#967 Inflation and distribution specs drastically improved
  - [x/gov] \#1773 Votes on a proposal can now be queried
  - [x/gov] Initial governance parameters can now be set in the genesis file
  - [x/stake] \#1815 Sped up the processing of `EditValidator` txs. 
  - [config] \#1930 Transactions indexer indexes all tags by default.

* SDK 
  - [baseapp] \#1587 Allow any alphanumeric character in route
  - [baseapp] Allow any alphanumeric character in route
  - [tools] Remove `rm -rf vendor/` from `make get_vendor_deps`
  - [x/auth] Recover ErrorOutOfGas panic in order to set sdk.Result attributes correctly
  - [x/bank] Unit tests are now table-driven
  - [tests] Add tests to example apps in docs
  - [tests] Fixes ansible scripts to work with AWS too
  - [tests] \#1806 CLI tests are now behind the build flag 'cli_test', so go test works on a new repo

BUG FIXES

* Gaia CLI  (`gaiacli`)
  -  \#1766 Fixes bad example for keybase identity

* Gaia
  - \#1804 Fixes gen-tx genesis generation logic temporarily until upstream updates
  - \#1799 Fix `gaiad export`
  - \#1839 Fixed bug where intra-tx counter wasn't set correctly for genesis validators
  - [x/stake] \#1858 Fixed bug where the cliff validator was not updated correctly
  - [tests] \#1675 Fix non-deterministic `test_cover` 
  - [tests] \#1551 Fixed invalid LCD test JSON payload in `doIBCTransfer`
  - [basecoin] Fixes coin transaction failure and account query [discussion](https://forum.cosmos.network/t/unmarshalbinarybare-expected-to-read-prefix-bytes-75fbfab8-since-it-is-registered-concrete-but-got-0a141dfa/664/6)
  - [x/gov] \#1757 Fix VoteOption conversion to String


