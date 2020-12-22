<!--
order: 3
-->

# Keepers

The nameservice module uses the [BankKeeper](docs.cosmos.network/v0.39/modules/bank/02_keepers.html) keepers in the `bank` module to check balances. Review the `bank` module code to ensure that permissions are limited as required for your application.

<!-- because the nameservice module relies on auth, bank, staking, distribution, slashing, and supply do we need to mention all keepers in play here? -->
