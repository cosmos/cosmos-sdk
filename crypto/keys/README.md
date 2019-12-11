# Keys API

[![API Reference](https://godoc.org/github.com/cosmos/cosmos-sdk/crypto/keys?status.svg)](https://godoc.org/github.com/cosmos/cosmos-sdk/crypto/keys)


## The Keybase interface

The [Keybase](https://godoc.org/github.com/cosmos/cosmos-sdk/crypto/keys#Keybase) interface defines
the methods that a type needs to implement to be used as key storage backend. This package provides
few implementations out-of-the-box.

## Constructors

### New

The [New](https://godoc.org/github.com/cosmos/cosmos-sdk/crypto/keys#New) constructor returns
an on-disk implementation backed by LevelDB storage that has been the default implementation used by the SDK until v0.38.0.
Due to [security concerns](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-006-secret-store-replacement.md), we recommend to drop
it in favor of the `NewKeyring` or `NewKeyringFile` constructors. We strongly advise to migrate away from this function as **it may be removed in a future
release**.

### NewInMemory

The [NewInMemory](https://godoc.org/github.com/cosmos/cosmos-sdk/crypto/keys#NewInMemory) constructor returns
an implementation backed by an in-memory, goroutine-safe map that we've historically used for testing purposes or on-the-fly
key generation and we consider safe for the aforementioned use cases since the generated keys are discarded when the process
terminates or the type instance is garbage collected.

### NewKeyring

The [NewKeyring](https://godoc.org/github.com/cosmos/cosmos-sdk/crypto/keys#NewKeyring) constructor returns
an implementation backed by the [Keyring](https://github.com/99designs/keyring) library, whose aim is to provide a common
abstraction and uniform interface between secret stores available for Windows, macOS, and most GNU/Linux distributions.
The instance returned by this constructor will use the operating system's default credentials store, which will then handle
keys storage operations securely. 

### NewKeyringFile, NewTestKeyring

Both [NewKeyringFile](https://godoc.org/github.com/cosmos/cosmos-sdk/crypto/keys#NewKeyringFile) and
[NewTestKeyring](https://godoc.org/github.com/cosmos/cosmos-sdk/crypto/keys#NewTestKeyring) constructors return
on-disk implementations backed by the [Keyring](https://github.com/99designs/keyring) `file` backend.
Whilst `NewKeyringFile` returns a secure, encrypted file-based type that requires user's password in order to
function correctly, the implementation returned by `NewTestKeyring` stores keys information in clear text and **must be used
only for testing purposes**.

`NewKeyringFile` and `NewTestKeyring` store key files in the client home directory's `keyring`
and `keyring-test` subdirectories respectively.
