# Changelog

## 0.5.0 (March 2, 2017)

BREAKING CHANGES

- nano: moved to `_nano` now while we're having build issues
- bcrypt: moved to `keys/bcrypt`
- hd: moved to `keys/hd`; `BTC` added to some function names; other function cleanup
- keys/cryptostore: moved to `keys`, renamed to `keybase`, and completely refactored
- keys: moved BIP39 related code to `keys/words`

FEATURE

- `Address` is a type alias for `cmn.HexBytes`

BUG FIX

- PrivKey comparisons done in constant time

## 0.4.1 (October 27, 2017)

This release removes support for bcrypt as it was merged too soon without an upgrade plan
for existing keys.

REVERTS THE FOLLOWING COMMITS:

- Parameterize and lower bcrypt cost - dfc4cdd2d71513e4a9922d679c74f36357c4c862
- Upgrade keys to use bcrypt with salts (#38)  - 8e7f0e7701f92206679ad093d013b9b162427631

## 0.4.0 (October 27, 2017)

BREAKING CHANGES:

- `keys`: use bcrypt plus salt

FEATURES:

- add support for signing via Ledger Nano

IMPROVEMENTS:

- linting and comments

## 0.3.0 (September 22, 2017)

BREAKING CHANGES:

- Remove `cmd` and `keys/tx` packages altogether: move it to the cosmos-sdk
- `cryptostore.Generator` takes a secret 
- Remove `String()` from `Signature` interface

FEATURES:

- `keys`: add CRC16 error correcting code

IMPROVEMENTS:

- Allow no passwords on keys for development convenience


## 0.2.1 (June 21, 2017)

- Improve keys command
  - No password prompts in non-interactive mode (echo 'foobar' | keys new foo)
  - Added support for seed phrases
    - Seed phrase now returned on `keys new`
    - Add `keys restore` to restore private key from key phrase
    - Checksum to verify typos in the seed phrase (rather than just a useless key)
  - Add `keys delete` to remove a key if needed

## 0.2.0 (May 18, 2017)

BREAKING CHANGES:

- [hd] The following functions no longer take a `coin string` as argument: `ComputeAddress`, `AddrFromPubKeyBytes`, `ComputeAddressForPrivKey`, `ComputeWIF`, `WIFFromPrivKeyBytes`
- Changes to `PrivKey`, `PubKey`, and `Signature` (denoted `Xxx` below):
  - interfaces are renamed `XxxInner`, and are not for use outside the package, though they must be exposed for sake of serialization.
  - `Xxx` is now a struct that wraps the corresponding `XxxInner` interface

FEATURES:

- `github.com/tendermint/go-keys -> github.com/tendermint/go-crypto/keys` - command and lib for generating and managing encrypted keys
- [hd] New function `WIFFromPrivKeyBytes(privKeyBytes []byte, compress bool) string`
- Changes to `PrivKey`, `PubKey`, and `Signature` (denoted `Xxx` below):
  - Expose a new method `Unwrap() XxxInner` on the `Xxx` struct which returns the corresponding `XxxInner` interface
  - Expose a new method `Wrap() Xxx` on the `XxxInner` interface which returns the corresponding `Xxx` struct

IMPROVEMENTS:

- Update to use new `tmlibs` repository

## 0.1.0 (April 14, 2017)

Initial release

