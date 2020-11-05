// Package keys provides common key management API.
//
//
// The Keybase interface
//
// The Keybase interface defines the methods that a type needs to implement to be used
// as key storage backend. This package provides few implementations out-of-the-box.
//
// NewLegacy
//
// The NewLegacy constructor returns an on-disk implementation backed by LevelDB storage that has been
// the default implementation used by the SDK until v0.38.0. Due to security concerns, it is
// recommended to drop it in favor of the NewKeyring constructor as it will be removed in future releases.
//
// NewInMemory
//
// The NewInMemory constructor returns an implementation backed by an in-memory, goroutine-safe
// map that has historically been used for testing purposes or on-the-fly key generation as the
// generated keys are discarded when the process terminates or the type instance is garbage
// collected.
//
// New
//
// The New constructor returns an implementation backed by a keyring library
// (https://github.com/99designs/keyring), whose aim is to provide a common abstraction and uniform
// interface between secret stores available for Windows, macOS, and most GNU/Linux distributions
// as well as operating system-agnostic encrypted file-based backends.
//
// The backends:
// 	os		The instance returned by this constructor uses the operating system's default
// 			credentials store to handle keys storage operations securely. It should be noted
// 			that the keyring keyring may be kept unlocked for the whole duration of the user
// 			session.
// 	file	This backend more closely resembles the previous keyring storage used prior to
// 			v0.38.1. It stores the keyring encrypted within the app's configuration directory.
// 			This keyring will request a password each time it is accessed, which may occur
// 			multiple times in a single command resulting in repeated password prompts.
// 	kwallet	This backend uses KDE Wallet Manager as a credentials management application:
// 			https://github.com/KDE/kwallet
// 	pass	This backend uses the pass command line utility to store and retrieve keys:
// 			https://www.passwordstore.org/
// 	test	This backend stores keys insecurely to disk. It does not prompt for a password to
// 			be unlocked and it should be use only for testing purposes.
// 	memory	Same instance as returned by NewInMemory. This backend uses a transient storage. Keys
// 			are discarded when the process terminates or the type instance is garbage collected.
package keyring
