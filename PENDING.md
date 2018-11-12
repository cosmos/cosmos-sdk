## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  * [cli] [\#2727](https://github.com/cosmos/cosmos-sdk/pull/2727) Fix unbonding command flow
  * [cli] [\#2786](https://github.com/cosmos/cosmos-sdk/pull/2786) Fix redelegation command flow

* Gaia

* SDK
  * [\#2752](https://github.com/cosmos/cosmos-sdk/pull/2752) Don't hardcode bondable denom.

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
    * [stake][cli] [\#2027] Add CLI query command for getting all delegations to a specific validator.
    
* Gaia

* SDK
    * [simulator] \#2682 MsgEditValidator now looks at the validator's max rate, thus it now succeeds a significant portion of the time

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  * [\#2749](https://github.com/cosmos/cosmos-sdk/pull/2749) Add --chain-id flag to gaiad testnet

* Gaia
 - #2672 [Makefile] Updated for better Windows compatibility and ledger support logic, get_tools was rewritten as a cross-compatible Makefile.

* SDK
 - [x/mock/simulation] [\#2720] major cleanup, introduction of helper objects, reorganization

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia
  * [\#2723] Use `cosmosvalcons` Bech32 prefix in `tendermint show-address`
  * [\#2742](https://github.com/cosmos/cosmos-sdk/issues/2742) Fix time format of TimeoutCommit override

* SDK

* Tendermint
  * [\#2797](https://github.com/tendermint/tendermint/pull/2797) AddressBook requires addresses to have IDs; Do not crap out immediately after sending pex addrs in seed mode
