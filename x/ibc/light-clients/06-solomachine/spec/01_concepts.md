<!--
order: 1
-->

# Concepts

## Proofs

A solo machine proof should verify that the solomachine public key signed 
over some specified data. The format for generating marshaled proofs for
the SDK's implementation of solo machine is as follows:

Construct the data using the associated protobuf definition and marshal it.

For example:
```go
data := &ClientStateData{
	Path:        []byte(path.String()),
	ClientState: any,
}

dataBz, err := cdc.MarshalBinaryBare(data)
```

Construct the `SignBytes` and marshal it.

For example:
```go
signBytes := &SignBytes{
	Sequence:    sequence,
	Timestamp:   timestamp,
	Diversifier: diversifier,
	Data:        dataBz,
}

signBz, err := cdc.MarshalBinaryBare(signBytes)
```

The helper functions in [proofs.go](../types/proofs.go) handle the above actions.

Sign the sign bytes. Embed the signatures into either `SingleSignatureData` or
`MultiSignatureData`. Convert the `SignatureData` to proto and marshal it. 

For example:
```go
sig, err := key.Sign(signBz)
sigData := &signing.SingleSignatureData{
	Signature: sig,
}

protoSigData := signing.SignatureDataToProto(sigData)
bz, err := cdc.MarshalBinaryBare(protoSigData)
```

Construct a `TimestampedSignature` and marshal it. The marshaled result can be
passed in as the proof parameter to the verification functions.

For example:
```go
timestampedSignature := &types.TimestampedSignature{
	Signature: sig,
	Timestamp: solomachine.Time,
}

proof, err := cdc.MarshalBinaryBare(timestampedSignature)
```
