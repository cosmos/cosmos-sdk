<!--
order: 4
-->
# Keyring Migrate Quick Start

`keyring` is the Cosmos SDK mechanism to manage the public/private keypair. Cosmos SDK v0.42 (Stargate) introduced breaking changes in the keyring. 

To upgrade your chain from v0.39 and earlier (Launchpad) to Stargate, you must migrate your keys inside the keyring to the latest version. For details on configuring and using the keyring, see [Setting up the keyring](../run-node/keyring.md).

This guide describes how to perform the keyring migration process.

The migration process is handled by the following CLI command:

```bash
Usage
simd keys migrate <old_home_dir>
```

The migration process moves key information from the legacy db-based Keybase to the [keyring](https://github.com/99designs/keyring)-based Keyring. The legacy Keybase persists keys in a LevelDB database in a 'keys' sub-directory of the client application home directory (`old_home_dir`). For example, `$HOME/.gaiacli/keys/` for [Gaia](https://github.com/cosmos/gaia).

For each key material entry, the command will prompt if the key should be skipped or not. If the key is not to be skipped, the passphrase must be entered. The key will only be migrated if the passphrase is correct. Otherwise, the command will exit and migration must be repeated.

The `migrate` CLI commands takes the following flags:
- `--dry-run` boolean flag. If it is set to false, it runs the migration without actually persisting any changes to the new Keybase. If it is set to true, it persists keys. This flag is useful for testing purposes: we recommend you to dry run the migration once before running it persistently.
- `--keyring-backend` string flag. It allows you to select a backend. For more detailed information about the available backends, you can read [the keyring guide](../run-node/keyring.md).
