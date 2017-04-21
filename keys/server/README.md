# Proxy Server

This package provides all the functionality for a local http server, providing access to key management functionality (creating, listing, updating, and deleting keys).  This is a nice building block for larger apps, and the HTTP handlers here can be embedded in a larger server that does nice things like signing transactions and posting them to a tendermint chain (which requires domain-knowledge of the transactions types and out of scope of this generic app).

## Key Management

We expose a number of methods for safely managing your keychain. If you are embedding this in a larger server, you will typically want to mount all these paths under `/keys`.

* `POST /` - provide a name and passphrase and create a brand new key
* `GET /` - get a list of all available key names, along with their public key and address
* `GET /{name}` - get public key and address for this named key
* `PUT /{name}` - update the passphrase for the given key. requires you to correctly provide the current passphrase, as well as a new one.
* `DELETE /{name}` - permanently delete this private key. requires you to correctly provide the current passphrase
