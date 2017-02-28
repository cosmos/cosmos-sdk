# Proxy Server

This package provides all the functionality for a local proxy http server, and ties together functionality from all the other packages to acheive this aim. Simply configure this server with your application-specific settings via a main script and you are good to go.

This server should run on the client's machine, and can accept a trusted connection from localhost (via tcp or unix socket).  It provides a simple json rest API and handles all the binary wrangling and cryptographic proofs under the hood.  Thus, you can host a local webapp (via electron?) and connect to this proxy, perform simple queries and posts and behind the scenes take advantage of the *awesome power* of the tendermint blockchain.

If you are writing native code, you can use this as well, or you can look for bindings to embed this functionality directly as a library in your codebase.
**(coming soon)**

## API

The API has various sections based on functionality.  The major portions are key management, signing and posting transactions, and querying and proving data.

### Key Management

We expose a number of methods for safely managing your keychain. They are typically bound under `/keys`, but could be placed in another location by the app.

* `POST /keys/` - provide a name and passphrase and create a brand new key
* `GET /keys/` - get a list of all available key names, along with their public key and address
* `GET /keys/{name}` - get public key and address for this named key

Later expose:

* `PUT /keys/{name}` - update the passphrase for the given key. requires you to correctly provide the current passphrase, as well as a new one.
* `DELETE /keys/{name}` - permanently delete this private key. requires you to correctly provide the current passphrase
* export and import functionality

### Transactions

You want to post your transaction.  Great.  Your application must provide logic to transform json into a `Signable` go struct.  Then we handle the rest, signing it with a keypair of your choosing, posting it to tendermint core, and returning you confirmation when it was committed.

* `POST /txs/` - provide name, passphrase and application-specific data to post to tendermint


### Proving Data

We sent some money to our friend, now we want to check his balance.  No, not just look at it, but really check it, verify all those cryptographic proofs that some node is not lying and it really, truly is in his account.

Thankfully, we have the technology and can do all the checks in the proxy, it might just take a second or two for us to get all those block headers.

However, this still just leaves us with a bunch of binary blobs from the server, so to make this whole process less painless, you should provide some application-specific logic to parse this binary data from the blockchain, so we can present it as json over the interface.

* `GET /query/{path}/{data}` - will quickly query the data under the given (hex-encoded) key.  `path` is `key` to query by merkle key, but your application could provide other prefixes, to differentiate by types (eg. `account`, `votes`, `escrow`).  The returned data is parsed into json and displayed.
* `GET /proof/{key}` - will query for a merkle proof of the given key, download block headers, and verify all the signatures of that block.  After it is done, it will present you some json and a stamp that it your data is really safe and sound.

## Configuring

When you instantiate a server, make sure to pass in application-specific info in order to properly. Like the following info:

Possibly as command-line flags:

* Where to store the private keys? (or find existing ones)
* Which type of key to generate?
* What is the URL of the tendermint RPC server?
  * TODO: support multiple node URLs and round-robin
* What is the chain_id we wish to connect to?

Extra code (plugin) you must write:

* Logic to parse json -> `Signable` transaction
* Logic to parse binary values from merkle tree -> `struct`ured data to render

TODO:

* How to get the trusted validator set?
