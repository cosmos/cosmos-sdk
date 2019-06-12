# ICS 23: Commitment

Package `commitment` defines types and methods to verify other chain's state. The main type is `Store`, containing 
proofs that can be verified when the correct value is provided. The spec functions those are directly related to 
verification are:

## Spec

```typescript
type verifyMembership = (root: CommitmentRoot, proof: CommitmentProof, key: Key, value: Value) => bool
type verifyNonMembership = (root: CommitmentRoot, proof: CommitmentProof, key: Key) => bool
```

## Impl

### types.go

`type Proof` implements `spec: type CommitmentProof`. CommitmentProof is an arbitrary object which can be used as
an argument for `spec: verifyMembership` / `spec: verifyNonMembership`, constructed with `spec: createMembershipProof` / 
`spec: createNonMembershipProof`. The implementation type `Proof` defines `spec: verify(Non)Membership` as its method 
`Verify(Root, []byte) error`, which takes the commitment root and the value bytes argument. The method acts as 
`spec: verifyMembership` when the value bytes is not nil, and `spec: verifyNonMembership` if it is nil.

`type Root` implements `spec: type CommitmentRoot`. 

In Cosmos-SDK implementation, `Root` will be the `AppHash []byte`, and `Proof` will be `merkle.Proof`, which consists
of `SimpleProof` and `IAVLValueProof`. Defined in `merkle/`

### store.go

`Store` assumes that the keys are already known at the time when the transaction is included, so the type `Proof` has
the method `Key() []byte`. The values should also have to be provided in order to verify the proof, but to reduce the
size of the transaction, they are excluded from `Proof` and provided by the application on runtime.

`NewStore` takes `[]Proof` as its argument, without verifying, since the values are yet unknown. They are stored in
`store.proofs`.

Proofs can be verified with `store.Prove()` method which takes the key of the proof it will verify and the value
that will be given to the `proof.Verify()`. Verified proofs are stored in `store.verified`.

### context.go

All of the ICS internals that requires verification on other chains' state are expected to take `ctx sdk.Context`
argument initialized by `WithStore()`. `WithStore()` sets the `Store` that contains the proofs for the other chain
in the context. Any attept to verify other chain's state without setting `Store` will lead to panic.

### value.go

Types in `value.go` is a replication of `store/mapping/*.go`, but only with a single method 
`Is(ctx sdk.Context, value T) bool`, which access on the underlying `Store` and performs verification.
