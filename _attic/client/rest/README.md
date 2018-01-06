## basecoin-server

### Proxy server
This package exposes access to key management i.e
- creating
- listing
- updating
- deleting

The HTTP handlers can be embedded in a larger server that
does things like signing transactions and posting them to a
Tendermint chain (which requires domain-knowledge of the transaction
types and is out of scope of this generic app).

### Key Management
We expose a couple of methods for safely managing your keychain.
If you are embedding this in a larger server, you will typically
want to mount all these paths /keys.

HTTP Method | Route | Description
---|---|---
POST|/|Requires a name and passphrase to create a brand new key
GET|/|Retrieves the list of all available key names, along with their public key and address
GET|/{name} | Updates the passphrase for the given key. It requires you to correctly provide the current passphrase, as well as a new one.
DELETE|/{name} | Permanently delete this private key. It requires you to correctly provide the current passphrase.
