<!--
order: 1
-->

# Concepts

## Client State

The `ClientState` for a solo machine light client stores the latest sequence, the frozen sequence,
the latest consensus state, and client flag indicating if the client should be allowed to be updated
after a governance proposal. 

If the client is not frozen then the frozen sequence is 0. 

## Consensus State

The consensus states stores the public key, diversifier, and timestamp of the solo machine light client. 

The diversifier is used to prevent accidental misbehaviour if the same public key is used across
different chains with the same client identifier. It should be unique to the chain the light client
is used on. 

## Public Key

The public key can be a single public key or a multi-signature public key. The public key type used
must fulfill the tendermint public key interface (this will become the SDK public key interface in the
near future). The public key must be registered on the application codec otherwise encoding/decoding 
errors will arise. The public key stored in the consensus state is represented as a protobuf `Any`. 
This allows for flexibility in what other public key types can be supported in the future. 
 
## Counterparty Verification

The solo machine light client can verify counterparty client state, consensus state, connection state,
channel state, packet commitments, packet acknowledgements, packet receipt absence, 
and the next sequence receive. At the end of each successful verification call the light
client sequence number will be incremented. 

Successful verification requires the current public key to sign over the proof.

## Proofs

A solo machine proof should verify that the solomachine public key signed
over some specified data. The format for generating marshaled proofs for
the SDK's implementation of solo machine is as follows:

1. Construct the data using the associated protobuf definition and marshal it.

For example:

```go
data := &ClientStateData{
  Path:        []byte(path.String()),
  ClientState: any,
}

dataBz, err := cdc.MarshalBinaryBare(data)
```

The helper functions `...DataBytes()` in [proofs.go](../types/proofs.go) handle this
functionality. 

2. Construct the `SignBytes` and marshal it.

For example:

```go
signBytes := &SignBytes{
  Sequence:    sequence,
  Timestamp:   timestamp,
  Diversifier: diversifier,
  DataType:    CLIENT,
  Data:        dataBz,
}

signBz, err := cdc.MarshalBinaryBare(signBytes)
```

The helper functions `...SignBytes()` in [proofs.go](../types/proofs.go) handle this functionality.
The `DataType` field is used to disambiguate what type of data was signed to prevent potential 
proto encoding overlap.

3. Sign the sign bytes. Embed the signatures into either `SingleSignatureData` or `MultiSignatureData`.
Convert the `SignatureData` to proto and marshal it.

For example:

```go
sig, err := key.Sign(signBz)
sigData := &signing.SingleSignatureData{
  Signature: sig,
}

protoSigData := signing.SignatureDataToProto(sigData)
bz, err := cdc.MarshalBinaryBare(protoSigData)
```

4. Construct a `TimestampedSignatureData` and marshal it. The marshaled result can be passed in 
as the proof parameter to the verification functions.

For example:

```go
timestampedSignatureData := &types.TimestampedSignatureData{
  SignatureData: sigData,
  Timestamp: solomachine.Time,
}

proof, err := cdc.MarshalBinaryBare(timestampedSignatureData)
```

## Updates By Header

An update by a header will only succeed if:

- the header provided is parseable to solo machine header
- the header sequence matches the current sequence
- the header timestamp is greater than or equal to the consensus state timestamp
- the currently registered public key generated the proof

If the update is successful:

- the public key is updated
- the diversifier is updated
- the timestamp is updated
- the sequence is incremented by 1
- the new consensus state is set in the client state 

## Updates By Proposal

An update by a governance proposal will only succeed if:

- the header provided is parseable to solo machine header
- the `AllowUpdateAfterProposal` client parameter is set to `true`
- the new header public key does not equal the consensus state public key

If the update is successful:

- the public key is updated
- the diversifier is updated
- the timestamp is updated
- the sequence is updated
- the new consensus state is set in the client state 
- the client is unfrozen (if it was previously frozen)

## Misbehaviour

Misbehaviour handling will only succeed if:

- the misbehaviour provided is parseable to solo machine misbehaviour
- the client is not already frozen
- the current public key signed over two unique data messages at the same sequence and diversifier. 

If the misbehaviour is successfully processed:

- the client is frozen by setting the frozen sequence to the misbehaviour sequence

NOTE: Misbehaviour processing is data processing order dependent. A misbehaving solo machine
could update to a new public key to prevent being frozen before misbehaviour is submitted. 

## Upgrades

Upgrades to solo machine light clients are not supported since an entirely different type of 
public key can be set using normal client updates.
