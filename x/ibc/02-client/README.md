# ICS 02: Client

Package `client` defines types and method to store and update light clients which tracks on other chain's state.
The main type is `Client`, which provides `commitment.Root` to verify state proofs and `ConsensusState` to 
verify header proofs.

## Spec

```typescript
interface ConsensusState {
  height: uint64
  root: CommitmentRoot
  validityPredicate: ValidityPredicate
  eqivocationPredicate: EquivocationPredicate
}

interface ClientState {
  consensusState: ConsensusState
  verifiedRoots: Map<uint64, CommitmentRoot>
  frozen: bool
}

interface Header {
  height: uint64
  proof: HeaderProof
  state: Maybe[ConsensusState]
  root: CommitmentRoot
}

type ValidityPredicate = (ConsensusState, Header) => Error | ConsensusState

type EquivocationPredicate = (ConsensusState, Header, Header) => bool
```

## Impl

### types.go

`spec: interface ConsensusState` is implemented by `type ConsensusState`. `ConsensusState.{GetHeight(), GetRoot(),
Validate(), Equivocation()}` each corresponds to `spec: ConsensusState.{height, root, validityPredicate, 
equivocationPredicate}`. `ConsensusState.Kind()` returns `Kind`, which is an enum indicating the type of the 
consensus algorithm.

`spec: interface Header` is implemented by `type Header`. `Header{GetHeight(), Proof(), State(), GetRoot()}` 
each corresponds to `spec: Header.{height, proof, state, root}`.

### manager.go

`spec: interface ClientState` is implemented by `type Object`. // TODO
