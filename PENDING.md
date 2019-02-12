## PENDING

BREAKING CHANGES

* Gaia REST API

* Gaia CLI

* Gaia

* SDK
 * \#3592 Drop deprecated keybase implementation's
   New constructor in favor of a new
   crypto/keys.New(string, string) implementation that
   returns a lazy keybase instance. Remove client.MockKeyBase,
   superseded by crypto/keys.NewInMemory()

* Tendermint

FEATURES

* Gaia REST API

* Gaia CLI

* Gaia

* SDK

* Tendermint


IMPROVEMENTS

* Gaia REST API

* Gaia CLI

* Gaia

* SDK
  * [\#3311] Reconcile the `DecCoin/s` API with the `Coin/s` API.
  * [\#3614] Add coin denom length checks to the coins constructors.
  * [\#3604] Improve SDK funds related error messages and allow for unicode in
  JSON ABCI log.

* Tendermint


BUG FIXES

* Gaia REST API

* Gaia CLI
  * [\#3586](https://github.com/cosmos/cosmos-sdk/pull/3586) Incomplete ledger derivation paths in keybase 

* Gaia
  * [\#3585] Fix setting the tx hash in `NewResponseFormatBroadcastTxCommit`.

* SDK
  * [\#3582](https://github.com/cosmos/cosmos-sdk/pull/3582) Running `make test_unit was failing due to a missing tag
  * [\#3617] Fix fee comparison when the required fees does not contain any denom
  present in the tx fees.

* Tendermint
