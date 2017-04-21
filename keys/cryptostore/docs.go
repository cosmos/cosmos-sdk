/*
package cryptostore maintains everything needed for doing public-key signing and
key management in software, based on the go-crypto library from tendermint.

It is flexible, and allows the user to provide a key generation algorithm
(currently Ed25519 or Secp256k1), an encoder to passphrase-encrypt our keys
when storing them (currently SecretBox from NaCl), and a method to persist
the keys (currently FileStorage like ssh, or MemStorage for tests).
It should be relatively simple to write your own implementation of these
interfaces to match your specific security requirements.

Note that the private keys are never exposed outside the package, and the
interface of Manager could be implemented by an HSM in the future for
enhanced security.  It would require a completely different implementation
however.

This Manager aims to implement Signer and KeyManager interfaces, along
with some extensions to allow importing/exporting keys and updating the
passphrase.

Encoder and Generator implementations are currently in this package,
keys.Storage implementations exist as subpackages of
keys/storage
*/
package cryptostore
