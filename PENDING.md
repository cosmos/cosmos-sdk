# PENDING CHANGELOG

<!----------------------------- BREAKING CHANGES ----------------------------->

## BREAKING CHANGES

### Gaia REST API

* [\#3641] Remove the ability to use a Keybase from the REST API client:
  * `password` and `generate_only` have been removed from the `base_req` object
  * All txs that used to sign or use the Keybase now only generate the tx
  * `keys` routes completely removed
* [\#3692] Update tx encoding and broadcasting endpoints:
  * Remove duplicate broadcasting endpoints in favor of POST @ `/txs`
    * The `Tx` field now accepts a `StdTx` and not raw tx bytes
  * Move encoding endpoint to `/txs/encode`

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

### Tendermint

<!------------------------------- IMPROVEMENTS ------------------------------->

## IMPROVEMENTS

### Gaia REST API

* [\#3692] Update tx encoding and broadcasting endpoints:
  * Remove duplicate broadcasting endpoints in favor of POST @ `/txs`
  * Move encoding endpoint to `/txs/encode`

### Gaia CLI

### Gaia

### SDK

* [\#3665] Overhaul sdk.Uint type in preparation for Coins's Int -> Uint migration.

### Tendermint

<!--------------------------------- BUG FIXES -------------------------------->

## BUG FIXES

### Gaia REST API

### Gaia CLI

### Gaia

### SDK

### Tendermint
