## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia
  * [gaiad init] \#2602 New genesis workflow

* SDK
  * [simulation] \#2665 only argument to simulation.Invariant is now app

* Tendermint
  * Upgrade to version 0.26.0

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

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia
 - #2637 [x/gov] Switched inactive and active proposal queues to an iterator based queue

* SDK
 - \#2573 [x/distribution] add accum invariance
 - \#2556 [x/mock/simulation] Fix debugging output
 - \#2396 [x/mock/simulation] Change parameters to get more slashes
 - \#2617 [x/mock/simulation] Randomize all genesis parameters
 - \#2669 [x/stake] Added invarant check to make sure validator's power aligns with its spot in the power store.
 - \#1924 [x/mock/simulation] Use a transition matrix for block size
 - \#2660 [x/mock/simulation] Staking transactions get tested far more frequently
 - \#2610 [x/stake] Block redelegation to and from the same validator
 - \#2652 [x/auth] Add benchmark for get and set account
 - \#2685 [store] Add general merkle absence proof (also for empty substores)
 - \#2708 [store] Disallow setting nil values

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia
 - \#2670 [x/stake] fixed incorrect `IterateBondedValidators` and split into two functions: `IterateBondedValidators` and `IterateLastBlockConsValidators`
 - \#2691 Fix local testnet creation by using a single canonical genesis time

* SDK
 - \#2625 [x/gov] fix AppendTag function usage error
 - \#2677 [x/stake, x/distribution] various staking/distribution fixes as found by the simulator
 - \#2674 [types] Fix coin.IsLT() impl, coins.IsLT() impl, and renamed coins.Is\* to coins.IsAll\* (see \#2686)
 - \#2711 [x/stake] Add commission data to `MsgCreateValidator` signature bytes.

* Tendermint
