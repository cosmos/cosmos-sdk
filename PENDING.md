## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  * [cli] [\#2595](https://github.com/cosmos/cosmos-sdk/issues/2595) Remove `keys new` in favor of `keys add` incorporating existing functionality with addition of key recovery functionality.
  * [cli] [\#2987](https://github.com/cosmos/cosmos-sdk/pull/2987) Add shorthand `-a` to `gaiacli keys show` and update docs
  * [cli] [\#2971](https://github.com/cosmos/cosmos-sdk/pull/2971) Additional verification when running `gaiad gentx`
  * [cli] [\#2734](https://github.com/cosmos/cosmos-sdk/issues/2734) Rewrite `gaiacli config`. It is now a non-interactive config utility.

* Gaia
 - [#128](https://github.com/tendermint/devops/issues/128) Updated CircleCI job to trigger website build on every push to master/develop.
 - [\#2994](https://github.com/cosmos/cosmos-sdk/pull/2994) Change wrong-password error message.
 - \#3009 Added missing Gaia genesis verification
 - [gas] \#3052 Updated gas costs to more reasonable numbers

* SDK
 - [auth] \#2952 Signatures are no longer serialized on chain with the account number and sequence number

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  - [\#2961](https://github.com/cosmos/cosmos-sdk/issues/2961) Add --force flag to gaiacli keys delete command to skip passphrase check and force key deletion unconditionally.

* Gaia

* SDK

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  * \#2991 Fully validate transaction signatures during `gaiacli tx sign --validate-signatures`

* Gaia

* SDK
 - \#1277 Complete bank module specification
 - \#2963 Complete auth module specification
  * \#2914 No longer withdraw validator rewards on bond/unbond, but rather move
  the rewards to the respective validator's pools.

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
  * [\#2921](https://github.com/cosmos/cosmos-sdk/issues/2921) Fix `keys delete` inability to delete offline and ledger keys.

* Gaia
  * [\#3003](https://github.com/cosmos/cosmos-sdk/issues/3003) CollectStdTxs() must validate DelegatorAddr against genesis accounts.

* SDK
  * \#2967 Change ordering of `mint.BeginBlocker` and `distr.BeginBlocker`, recalculate inflation each block

* Tendermint
