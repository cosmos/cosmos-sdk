## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  * [cli] [\#2595](https://github.com/cosmos/cosmos-sdk/issues/2595) Remove `keys new` in favor of `keys add` incorporating existing functionality with addition of key recovery functionality.

* Gaia
 - [#128](https://github.com/tendermint/devops/issues/128) Updated CircleCI job to trigger website build on every push to master/develop.
* SDK

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK
 - [x/mock/simulation] \#2832, \#2885, \#2873, \#2902 Simulation cleanup
 - [x/mock/simulation] [\#2720] major cleanup, introduction of helper objects, reorganization
 - \#2821 Codespaces are now strings
 - [types] #2776 Improve safety of `Coin` and `Coins` types. Various functions
 and methods will panic when a negative amount is discovered.
 - #2815 Gas unit fields changed from `int64` to `uint64`.
 - #2821 Codespaces are now strings
 - #2779 Introduce `ValidateBasic` to the `Tx` interface and call it in the ante
 handler.
 - #2825 More staking and distribution invariants
 - #2912 Print commit ID in hex when commit is synced.
 * Use `CodeInternal` instead of `CodeInvalidSequence` in the ante handler when
 an invalid account number is given.
 - \#1277 Complete bank module specification
 
* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  * [\#2921](https://github.com/cosmos/cosmos-sdk/issues/2921) Fix `keys delete` inability to delete offline and ledger keys.

* Gaia

* SDK

* Tendermint
