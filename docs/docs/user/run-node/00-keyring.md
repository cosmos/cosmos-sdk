---
sidebar_position: 1
---

# Setting up the keyring

:::note Synopsis
This document describes how to configure and use the keyring and its various backends for an [**application**](../../learn/beginner/00-app-anatomy.md).
:::

The keyring holds the private/public key pairs used to interact with a node. For instance, a validator key needs to be set up before running the blockchain node, so that blocks can be correctly signed. The private key can be stored in different locations, called "backends," such as a file or the operating system's own key storage.

## Available backends for the keyring

Starting with the v0.38.0 release, Cosmos SDK comes with a new keyring implementation
that provides a set of commands to manage cryptographic keys in a secure fashion. The
new keyring supports multiple storage backends, some of which may not be available on
all operating systems.

### The `os` backend

The `os` backend relies on operating system-specific defaults to handle key storage
securely. Typically, an operating system's credential subsystem handles password prompts,
private keys storage, and user sessions according to the user's password policies. Here
is a list of the most popular operating systems and their respective password managers:

* macOS: [Keychain](https://support.apple.com/en-gb/guide/keychain-access/welcome/mac)
* Windows: [Credentials Management API](https://docs.microsoft.com/en-us/windows/win32/secauthn/credentials-management)
* GNU/Linux:
    * [libsecret](https://gitlab.gnome.org/GNOME/libsecret)
    * [kwallet](https://api.kde.org/frameworks/kwallet/html/index.html)
    * [keyctl](https://www.kernel.org/doc/html/latest/security/keys/core.html)

GNU/Linux distributions that use GNOME as the default desktop environment typically come with
[Seahorse](https://wiki.gnome.org/Apps/Seahorse). Users of KDE based distributions are
commonly provided with [KDE Wallet Manager](https://userbase.kde.org/KDE_Wallet_Manager).
Whilst the former is in fact a `libsecret` convenient frontend, the latter is a `kwallet`
client. `keyctl` is a secure backend that leverages the Linux kernel security key management system
to store cryptographic keys securely in memory.

`os` is the default option since operating system's default credentials managers are
designed to meet users' most common needs and provide them with a comfortable
experience without compromising on security.

The recommended backends for headless environments are `file` and `pass`.

### The `file` backend

The `file` backend more closely resembles the keybase implementation used prior to
v0.38.1. It stores the keyring encrypted within the app's configuration directory. This
keyring will request a password each time it is accessed, which may occur multiple
times in a single command resulting in repeated password prompts. If using bash scripts
to execute commands using the `file` option you may want to utilize the following format
for multiple prompts:

```shell
# assuming that KEYPASSWD is set in the environment
$ gaiacli config keyring-backend file                             # use file backend
$ (echo $KEYPASSWD; echo $KEYPASSWD) | gaiacli keys add me        # multiple prompts
$ echo $KEYPASSWD | gaiacli keys show me                          # single prompt
```

:::tip
The first time you add a key to an empty keyring, you will be prompted to type the password twice.
:::

### The `pass` backend

The `pass` backend uses the [pass](https://www.passwordstore.org/) utility to manage on-disk
encryption of keys' sensitive data and metadata. Keys are stored inside `gpg` encrypted files
within app-specific directories. `pass` is available for the most popular UNIX
operating systems as well as GNU/Linux distributions. Please refer to its manual page for
information on how to download and install it.

:::tip
**pass** uses [GnuPG](https://gnupg.org/) for encryption. `gpg` automatically invokes the `gpg-agent`
daemon upon execution, which handles the caching of GnuPG credentials. Please refer to `gpg-agent`
man page for more information on how to configure cache parameters such as credentials TTL and
passphrase expiration.
:::

The password store must be set up prior to first use:

```shell
pass init <GPG_KEY_ID>
```

Replace `<GPG_KEY_ID>` with your GPG key ID. You can use your personal GPG key or an alternative
one you may want to use specifically to encrypt the password store.

### The `kwallet` backend

The `kwallet` backend uses `KDE Wallet Manager`, which comes installed by default on the
GNU/Linux distributions that ship KDE as the default desktop environment. Please refer to
[KWallet API documentation](https://api.kde.org/frameworks/kwallet/html/index.html) for more
information.

### The `keyctl` backend

The *Kernel Key Retention Service* is a security facility that
has been added to the Linux kernel relatively recently. It allows sensitive
cryptographic data such as passwords, private key, authentication tokens, etc
to be stored securely in memory.

The `keyctl` backend is available on Linux platforms only.

### The `test` backend

The `test` backend is a password-less variation of the `file` backend. Keys are stored
unencrypted on disk.

**Provided for testing purposes only. The `test` backend is not recommended for use in production environments**.

### The `memory` backend

The `memory` backend stores keys in memory. The keys are immediately deleted after the program has exited.

**Provided for testing purposes only. The `memory` backend is not recommended for use in production environments**.

### Setting backend using an env variable 

You can set the keyring-backend using env variable: `BINNAME_KEYRING_BACKEND`. For example, if your binary name is `gaia-v5` then set: `export GAIA_V5_KEYRING_BACKEND=pass`

## Adding keys to the keyring

:::warning
Make sure you can build your own binary, and replace `simd` with the name of your binary in the snippets.
:::

Applications developed using the Cosmos SDK come with the `keys` subcommand. For the purpose of this tutorial, we're running the `simd` CLI, which is an application built using the Cosmos SDK for testing and educational purposes. For more information, see [`simapp`](https://github.com/cosmos/cosmos-sdk/tree/main/simapp).

You can use `simd keys` for help about the keys command and `simd keys [command] --help` for more information about a particular subcommand.

To create a new key in the keyring, run the `add` subcommand with a `<key_name>` argument. For the purpose of this tutorial, we will solely use the `test` backend, and call our new key `my_validator`. This key will be used in the next section.

```bash
$ simd keys add my_validator --keyring-backend test

# Put the generated address in a variable for later use.
MY_VALIDATOR_ADDRESS=$(simd keys show my_validator -a --keyring-backend test)
```

This command generates a new 24-word mnemonic phrase, persists it to the relevant backend, and outputs information about the keypair. If this keypair will be used to hold value-bearing tokens, be sure to write down the mnemonic phrase somewhere safe!

By default, the keyring generates a `secp256k1` keypair. The keyring also supports `ed25519` keys, which may be created by passing the `--algo ed25519` flag. A keyring can of course hold both types of keys simultaneously, and the Cosmos SDK's `x/auth` module supports natively these two public key algorithms.

### Post-quantum keys from a 24-word mnemonic

`simd keys add --key-type ml_dsa_65` derives ML-DSA-65 software account keys with the
draft PQHD derivation, inspired by the
[Project Eleven approach](https://www.projecteleven.com/blog/generating-post-quantum-keypairs-from-a-single-24word-seed-phrase):
BIP39 mnemonic → BIP32 master → fully hardened child → 64-byte BIP85 entropy → SHAKE256 → ML-DSA-65 key generation.

**Experimental:** This implements PQHD draft v0.1.0, not a finalized standard.
Do not protect real assets with keys derived under this draft.

```bash
# Generate a fresh 24-word mnemonic and store an ML-DSA-65 key.
simd keys add pq --key-type ml_dsa_65 --keyring-backend file

# Restore it in a new keyring using the same mnemonic.
simd keys add pq --key-type ml_dsa_65 --recover --keyring-backend file

# Derive another account from that SAME mnemonic (not a newly generated one).
simd keys add pq-next --key-type ml_dsa_65 --recover \
  --account 7 --index 9 --keyring-backend file
```

The implementation pins the merged [PQHD framework](https://github.com/pqdh/pqhd/blob/befdce14cecbb5d485d3239982bfa5b207cdb922/spec/README.md),
[family-block registry](https://github.com/pqdh/pqhd/blob/befdce14cecbb5d485d3239982bfa5b207cdb922/registry/registry.json),
and [ML-DSA-65 test vectors](https://github.com/pqdh/pqhd/blob/befdce14cecbb5d485d3239982bfa5b207cdb922/test-vectors/201.json)
at the same revision. This supersedes the earlier PR #2 scheme assignments, but remains pre-v1.
The registry assigns families blocks of 100 indices; block 0 is never assigned.
ML-DSA owns block 2, and its active ML-DSA-65 entry is `201`, with a 32-byte `xi` input
to `ML-DSA.KeyGen_internal`.

The path is `m/83696968'/505153'/coin_type'/201'/account'/index'`.
The default `simd` path is `m/83696968'/505153'/118'/201'/0'/0'`.

* `83696968'` is the BIP85 purpose.
* `505153'` is PQHD's proposed application namespace ("PQS"); scheme `201'` identifies ML-DSA-65.
  These are **not registered BIP85 or Project Eleven assignments**.
* `--coin-type` selects the SLIP-44 coin type and defaults to the application's configured value
  (`118` for Cosmos Hub). Different coin types derive different keys from the same mnemonic.
* Account and index default to zero. Coin type, account, and index must each be less than 2³¹.
  Every component is hardened; there is no BIP44 change level.
  An invalid BIP32 master or child causes an error, never a silent retry at the next index.
* `--hd-path` may override coin type/account/index, but must use this exact six-component namespace.
  Ledger devices do not support `ml_dsa_65` keys.

Only active scheme `201` is implemented here. Other schemes, including reserved FN-DSA
indices `400` and `401`, are rejected. PQHD requires deprecated entries to remain derivable
for recovery and discourages new keys under them; there are no deprecated entries in this
pinned registry. The obsolete draft ID `3` is **unassigned**, not a deprecated recovery alias.
The account-key API does not expose intermediate BIP32 xpubs or xprivs.

The 32-byte big-endian child private scalar is processed with
`HMAC-SHA512(key="bip-entropy-from-k", message=childPrivateKey)`.
**All 64 bytes** seed SHAKE256; its first 32 output bytes become ML-DSA-65's `xi`.
The HMAC is not truncated before expansion. No new signature implementation, wire format,
address scheme, or consensus rule is introduced.

**Recovery:** Back up the mnemonic, implementation revision, derivation mode (PQHD, or `--legacy-bip44`), coin type/account/index or full path,
and any BIP39 passphrase. To set or recover a nonempty BIP39 passphrase, also pass `--interactive`.
This passphrase is different from the keyring encryption password. The mnemonic must be a
checksum-valid 24-word English BIP39 sentence: it is NFKD-normalized, split on Unicode whitespace,
and rejoined with single ASCII spaces. The passphrase is NFKD-normalized **without trimming**;
leading/trailing spaces and tabs matter, including in the CLI. A wrong passphrase silently derives
a different valid key. Software keyring records store the resulting private key, not the mnemonic
or derivation metadata. Signing with a stored key does not depend on how it was derived.

**Opting out with `--legacy-bip44`:** Earlier releases derived `ml_dsa_65` keys on the BIP44
path `m/44'/coin_type'/account'/0/index`, shared with `secp256k1` keys. To recover such a key,
pass `--legacy-bip44`; without it, the same mnemonic derives a different key and address.
`--legacy-bip44` accepts shorter mnemonics and prints a warning, because the resulting key shares
the BIP44 index space with `secp256k1` keys. `secp256k1` derivation is unchanged. Switching
derivations does not migrate funds or replace an existing account's public key.

```bash
simd keys add old-pq --key-type ml_dsa_65 --legacy-bip44 --recover --keyring-backend file
```

Go code calling `keyring.NewAccount` with `hd.MlDsa65` keeps the BIP44 derivation. Use
`hd.MlDsa65BIP85` with a path from `hd.CreateMlDsa65BIP85Path` for PQHD.

**Restoring every account from a mnemonic:** A stored key does not record which derivation produced
it, and an address does not reveal it either: both derivation modes produce the same `ml_dsa_65`
public key type. Under PQHD, the scheme is part of the path, so a wallet can find a user's accounts
from the mnemonic alone by scanning a finite set of candidates:

1. For each PQHD scheme the wallet supports (currently only `201'`, ML-DSA-65), derive
   `m/83696968'/505153'/coin_type'/scheme'/account'/index'` for increasing account and index
   values, and stop each scan after a run of unused addresses (a gap limit, conventionally 20).
   A later registry version may add schemes under new IDs. Deprecated schemes stay derivable,
   so a wallet should keep scanning them for recovery.
2. For keys created before PQHD, also derive the BIP44 path `m/44'/coin_type'/account'/0/index`
   **twice**, once as `secp256k1` and once as legacy `ml_dsa_65`. On this path the mnemonic alone
   cannot tell the two apart.
3. Treat a candidate as an account in use if its address has an on-chain account or history.

Each PQHD candidate belongs to exactly one scheme, so step 1 never confuses algorithms.
Step 2 is the ambiguity that PQHD removes for new keys.

Go wallets can get the candidate list for one coin type, account, and index from
`hd.AccountKeyDerivations(coinType, account, index)`. Each entry names its scheme, public key
type, path, and the algorithm to pass to `keyring.NewAccount`. The list grows when a new scheme
is supported, so wallets that iterate over it pick up new schemes without code changes.

PQHD candidates require a 24-word mnemonic and return `hd.ErrMlDsa65BIP85Mnemonic` for a valid
shorter backup, such as a 12-word mnemonic from an earlier release. Skip only that error and keep
going, so the BIP44 candidates are still tried. Report any other error, such as a bad checksum:

```go
derivations, err := hd.AccountKeyDerivations(coinType, account, index)
if err != nil {
    return err
}
for _, d := range derivations {
    bz, err := d.Algo.Derive()(mnemonic, bip39Passphrase, d.Path)
    if errors.Is(err, hd.ErrMlDsa65BIP85Mnemonic) {
        continue // PQHD needs 24 words; shorter backups only have BIP44 keys.
    }
    if err != nil {
        return err
    }
    pubKey := d.Algo.Generate()(bz).PubKey()
    // Look up pubKey.Address() on chain; if it is in use, restore it with
    // kr.NewAccount(name, mnemonic, bip39Passphrase, d.Path, d.Algo).
}
```

**Keys from earlier unreleased PQHD drafts are incompatible.** The initial SDK-local path
(`m/83696968'/118'/65'/account'/index'`), the scoped derivation that used the first 32 HMAC
bytes directly, and the subsequent SHAKE256 draft using scheme `3` all produce different keys.
The old scoped path `m/83696968'/505153'/coin_type'/3'/account'/index'` is rejected;
there is no silent fallback to any earlier derivation.
Existing stored private keys still sign unchanged, but mnemonic recovery of those draft keys
requires the original implementation and derivation settings. Preserve those backups before
updating; this change does not migrate their funds.

Go applications can use the same flow through the existing keyring API:

```go
path, err := hd.CreateMlDsa65BIP85Path(118, 7, 9)
if err != nil {
    return err
}
record, err := kr.NewAccount("pq", mnemonic, bip39Passphrase, path, hd.MlDsa65BIP85)
```

Use the same coin type, account, index, mnemonic, and passphrase to reproduce a scoped
PQHD ML-DSA-65 key. These keys intentionally differ from Project Eleven's unscoped
namespace (`503131'`) and direct HMAC truncation.

For other applications, `hd.DeriveBIP85Entropy(masterPrivateKey, chainCode, path, numBytes)`
accepts BIP32 master material from `hd.ComputeMastersFromSeed` and returns 1–64 secret bytes.
This generic BIP85 API returns raw HMAC entropy, **not** PQHD's SHAKE256 output.
It supports the article's exact paths `m/83696968'/503131'/0'` (32 bytes for ML-DSA-44)
and `m/83696968'/503131'/2'` (48 bytes for SLH-DSA-SHA2-128s). However, **this integration
generates SDK account keys only for ML-DSA-65**; it does not add ML-DSA-44 or SLH-DSA signing,
key codecs, or transaction acceptance. Never reuse an entropy path for different algorithms.

:::warning
This draft derivation integration needs independent security review before production use.
It does not make an entire chain quantum-safe or strengthen a previously exposed mnemonic.
Do not export extended public or private keys for PQHD path nodes. Keep the mnemonic,
master/ancestor extended keys, and intermediate BIP32 keys secret: BIP32 still uses secp256k1,
and an exposed ancestor extended public key is vulnerable to a future quantum attacker.
Prefer a fresh mnemonic if an existing wallet's exposure is uncertain. The mnemonic and BIP32
scalars cap the derived keys at about 128 bits of quantum security, despite ML-DSA-65's
NIST category-3 classification.
Buffers are cleared where practical, but Go does not guarantee complete erasure of secret copies.
:::
