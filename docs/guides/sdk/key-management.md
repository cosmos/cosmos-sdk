# Key Management

Here we cover many aspects of handling keys within the Cosmos SDK
framework.

## Pseudo Code

Generating an address for an ed25519 public key (in pseudo code):

```
const TypeDistinguisher = HexToBytes("1624de6220")

// prepend the TypeDistinguisher as Bytes
SerializedBytes = TypeDistinguisher ++ PubKey.asBytes()

Address = ripemd160(SerializedBytes)
```
