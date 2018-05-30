# go-crypto [![GoDoc](https://godoc.org/github.com/tendermint/go-crypto?status.svg)](https://godoc.org/github.com/tendermint/go-crypto)

go-crypto is the cryptographic package adapted for Tendermint's uses

## Importing it
`import "github.com/tendermint/go-crypto"`

## Binary encoding

For Binary encoding, please refer to the [Tendermint encoding spec](https://github.com/tendermint/tendermint/blob/develop/docs/spec/blockchain/encoding.md).

## JSON Encoding

go-crypto `.Bytes()` uses Amino:binary encoding, but Amino:JSON is also supported.

```go
Example Amino:JSON encodings:

crypto.PrivKeyEd25519     - {"type":"954568A3288910","value":"EVkqJO/jIXp3rkASXfh9YnyToYXRXhBr6g9cQVxPFnQBP/5povV4HTjvsy530kybxKHwEi85iU8YL0qQhSYVoQ=="}
crypto.SignatureEd25519   - {"type":"6BF5903DA1DB28","value":"77sQNZOrf7ltExpf7AV1WaYPCHbyRLgjBsoWVzcduuLk+jIGmYk+s5R6Emm29p12HeiNAuhUJgdFGmwkpeGJCA=="}
crypto.PubKeyEd25519      - {"type":"AC26791624DE60","value":"AT/+aaL1eB0477Mud9JMm8Sh8BIvOYlPGC9KkIUmFaE="}
crypto.PrivKeySecp256k1   - {"type":"019E82E1B0F798","value":"zx4Pnh67N+g2V+5vZbQzEyRerX9c4ccNZOVzM9RvJ0Y="}
crypto.SignatureSecp256k1 - {"type":"6D1EA416E1FEE8","value":"MEUCIQCIg5TqS1l7I+MKTrSPIuUN2+4m5tA29dcauqn3NhEJ2wIgICaZ+lgRc5aOTVahU/XoLopXKn8BZcl0bnuYWLvohR8="}
crypto.PubKeySecp256k1    - {"type":"F8CCEAEB5AE980","value":"A8lPKJXcNl5VHt1FK8a244K9EJuS4WX1hFBnwisi0IJx"}
```
