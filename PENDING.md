# PENDING CHANGELOG

<!----------------------------- BREAKING CHANGES ----------------------------->

## BREAKING CHANGES

### Gaia REST API

### Gaia CLI

### Gaia

### SDK

### Tendermint

<!--------------------------------- FEATURES --------------------------------->

## FEATURES

### Gaia REST API

### Gaia CLI

### Gaia

### SDK

* [\3813](https://github.com/cosmos/cosmos-sdk/pull/3813) New sdk.NewCoins safe constructor to replace bare
  sdk.Coins{} declarations.

### Tendermint

<!------------------------------- IMPROVEMENTS ------------------------------->

## IMPROVEMENTS

### Gaia REST API

### Gaia CLI

* [\#3841](https://github.com/cosmos/cosmos-sdk/pull/3841) Add indent to JSON of `gaiacli keys [add|show|list]`
* [\#3859](https://github.com/cosmos/cosmos-sdk/pull/3859) Add newline to echo of `gaiacli keys ...`

### Gaia

* #3808 `gaiad` and `gaiacli` integration tests use ./build/ binaries.

### SDK

* [\#3820] Make Coins.IsAllGT() more robust and consistent.
* [\#3864] Make Coins.IsAllGTE() more consistent.

* #3801 `baseapp` saftey improvements

### Tendermint

### CI/CD

* [\198](https://github.com/cosmos/cosmos-sdk/pull/3832)

<!--------------------------------- BUG FIXES -------------------------------->

## BUG FIXES

### Gaia REST API

### Gaia CLI

### Gaia

### SDK

* [\#3837] Fix `WithdrawValidatorCommission` to properly set the validator's
remaining commission.

### Tendermint
