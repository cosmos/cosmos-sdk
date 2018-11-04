## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK
  * [simulation] \#2665 only argument to simulation.Invariant is now app

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)
    * [cli] [\#2569](https://github.com/cosmos/cosmos-sdk/pull/2569) Add commands to query validator unbondings and redelegations
    * [cli] [\#2569](https://github.com/cosmos/cosmos-sdk/pull/2569) Add commands to query validator unbondings and redelegations
    * [cli] [\#2524](https://github.com/cosmos/cosmos-sdk/issues/2524) Add support offline mode to `gaiacli tx sign`. Lookups are not performed if the flag `--offline` is on.
    * [cli] [\#2558](https://github.com/cosmos/cosmos-sdk/issues/2558) Rename --print-sigs to --validate-signatures. It now performs a complete set of sanity checks and reports to the user. Also added --print-signature-only to print the signature only, not the whole transaction.

* Gaia

* SDK
    * (#1336) Mechanism for SDK Users to configure their own Bech32 prefixes instead of using the default cosmos prefixes.
    * [simulator] \#2682 MsgEditValidator now looks at the validator's max rate, thus it now succeeds a significant portion of the time

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK
 - #2573 [x/distribution] add accum invariance
 - \#1924 [x/mock/simulation] Use a transition matrix for block size
 - \#2660 [x/mock/simulation] Staking transactions get tested far more frequently
 - #2610 [x/stake] Block redelegation to and from the same validator
 - #2652 [x/auth] Add benchmark for get and set account

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK
 - #2625 [x/gov] fix AppendTag function usage error


* Tendermint
