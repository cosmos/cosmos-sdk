# Keyring Migrate Quick Start

`keyring` is mechanism of managing private/public keypair in cosmos-sdk. 

The more detailed information about keyring you can find here [here](../run-node/keyring.md)


## Backends

The `keyring` supports the following backends:

* `os` uses the operating system's default credentials store.
* `file` uses encrypted file-based keystore within the app's configuration directory. This keyring will request a password each time it is accessed, which may occur multiple times in a single command resulting in repeated password prompts.
* `kwallet` uses KDE Wallet Manager as a credentials management application.
* `pass` uses the pass command line utility to store and retrieve keys.
* `test` stores keys insecurely to disk. It does not prompt for a password to be unlocked and it should be use only for testing purposes.


## CLI command

Usage:
simd migrate <old_home_dir>

Migrates key information from the legacy (db-based) Keybase to the new keyring-based Keyring. The legacy Keybase used to persist keys in a LevelDB database stored in a 'keys' sub-directory of the old client application's home directory **old_home_dir**, e.g. $HOME/.gaiacli/keys/. For each key material entry, the command will prompt if the key should be skipped or not. If the key is not to be skipped, the passphrase must be entered. The key will only be migrated if the passphrase is correct. Otherwise, the command will exit and migration must be repeated.

You can find command implementation [here](../../client/keys/migrate.go)

## Flags
`FlagDryRun` bool flag. If it's set to false, it runs migration without actually persisting any changes to the new Keybase. If it's set to true, it persists keys. This flag is useful for testing purposes.



