## 2 Connections

([Back to table of contents](README.md#contents))

The basis of IBC is the ability to verify in the on-chain consensus ruleset of chain _B_ that a data packet received on chain _B_ was correctly generated on chain _A_. This establishes a cross-chain linearity guarantee: upon validation of that packet on chain _B_ we know that the packet has been executed on chain _A_ and any associated logic resolved (such as assets being escrowed), and we can safely perform application logic on chain _B_ (such as generating vouchers on chain _B_ for the chain _A_ assets which can later be redeemed with a packet in the opposite direction).

This section outlines the abstraction of an IBC _connection_: the state and consensus ruleset necessary to perform IBC packet verification. 

### 2.1 Definitions

- Chain _A_ is the source blockchain from which the IBC packet is sent
- Chain _B_ is the destination blockchain on which the IBC packet is received
- _H<sub>h</sub>_ is the signed header of chain _A_ at height _h_
- _C<sub>h</sub>_ is the consensus ruleset of chain _A_ at height _h_
- _V<sub>k,h</sub>_ is the value stored on chain _A_ under key _k_ at height _h_
- _P_ is the unbonding period of chain _A_, in units of time
- &#916;_(a, b)_ is the time difference between events _a_ and _b_

Note that of all these, only _H<sub>h</sub>_ defines a signature and is thus attributable.

### 2.2 Requirements

To facilitate an IBC connection, the two blockchains must provide the following proofs:

1. Given a trusted _H<sub>h</sub>_ and _C<sub>h</sub>_ and an attributable update message _U<sub>h'</sub>_,  
   it is possible to prove _H<sub>h'</sub>_ where _C<sub>h'</sub> = C<sub>h</sub>_ and &#916;_(now, H<sub>h</sub>) < P_
2. Given a trusted _H<sub>h</sub>_ and _C<sub>h</sub>_ and an attributable change message _X<sub>h'</sub>_,  
   it is possible to prove _H<sub>h'</sub>_ where _C<sub>h'</sub>_ &#8800; _C<sub>h</sub>_ and &#916; _(now, H<sub>h</sub>) < P_
3. Given a trusted _H<sub>h</sub>_ and a merkle proof _M<sub>k,v,h</sub>_ it is possible to prove _V<sub>k,h</sub>_

It is possible to make use of the structure of BFT consensus to construct extremely lightweight and provable messages _U<sub>h'</sub>_ and _X<sub>h'</sub>_. The implementation of these requirements with Tendermint consensus is defined in Appendix E. Another algorithm able to provide equally strong guarantees (such as Casper) is also compatible with IBC but must define its own set of update and change messages.

The merkle proof _M<sub>k,v,h</sub>_ is a well-defined concept in the blockchain space, and provides a compact proof that the key value pair (_k, v)_ is consistent with a merkle root stored in _H<sub>h</sub>_. Handling the case where _k_ is not in the store requires a separate proof of non-existence, which is not supported by all merkle stores. Thus, we define the proof only as a proof of existence. There is no valid proof for missing keys, and we design the algorithm to work without it.

_valid(H<sub>h </sub>,M<sub>k,v,h </sub>)_ &#8658; _[true | false]_

### 2.3 Connection Lifecycle

#### 2.3.1 Opening a Connection

All proofs require an initial _H<sub>h</sub>_ and _C<sub>h</sub>_ for some _h_, where &#916;_(now, H<sub>h</sub>) < P_.

Establishing a bidirectional initial root-of-trust between the two blockchains (_A_ to _B_ and _B_ to _A_) — _HA<sub>h</sub>_ and _CA<sub>h</sub>_ stored on chain _B_, and _HB<sub>h</sub>_ and _CB<sub>h</sub>_ stored on chain _A_ — is necessary before any IBC packets can be sent. 

Any header may be from a malicious chain (e.g. shadowing a real chain state with a fake validator set), so a subjective decision is required before establishing a connection. This can be performed permissionlessly, in which case users later utilizing the IBC channel must check the root-of-trust themselves, or authorized by on-chain governance for additional assurance.

#### 2.3.2 Following Block Headers

We define two messages _U<sub>h</sub>_ and _X<sub>h</sub>_, which together allow us to securely advance our trust from some known _H<sub>n</sub>_ to some future _H<sub>h</sub>_ where _h > n_. Some implementations may require that _h = n + 1_ (all headers must be processed in order). IBC implemented on top of Tendermint or similar BFT algorithms requires only that &#916;_<sub>vals</sub>(C<sub>n</sub>, C<sub>h</sub> ) < ⅓_ (each step must have a change of less than one-third of the validator set)[[4](./references.md#4)].

Either requirement is compatible with IBC. However, by supporting proofs where  _h_-_n > 1_, we can follow the block headers much more efficiently in situations where the majority of blocks do not include an IBC packet between chains _A_ and _B_, and enable low-bandwidth connections to be implemented at very low cost. If there are packets to relay every block, these two requirements collapse to the same case (every header must be relayed).

Since these messages _U<sub>h</sub>_ and _X<sub>h</sub>_ provide all knowledge of the remote blockchain, we require that they not just be provable, but also attributable. As such, any attempt to violate the finality guarantees in headers posted to chain _B_ can be submitted back to chain _A_ for punishment, in the same manner that chain _A_ would independently punish (slash) identified Byzantine actors.

More formally, given existing set of trust _T_ = _{(H<sub>i </sub>, C<sub>i </sub>), (H<sub>j </sub>, C<sub>j </sub>), …}_, we must provide:

_valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>)_ &#8658; _[true | false | unknown]_

_if H<sub>h-1</sub>_ &#8712; _T then_:
* _valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>)_ &#8658; _[true | false]_
* 	∃ (U<sub>h</sub> | X<sub>h</sub>)  &#8658; valid(T, X<sub>h</sub> | U<sub>h</sub>) {aren't there infinite? why is this necessary}

_if C<sub>h</sub>_ &#8713; _T then_
* _valid(T, U<sub>h </sub>)_ &#8658; _false_

We can then process update transactions as follows:

_update(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>)_  &#8658; match _valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>)_ with
* _false_ &#8658; fail with `invalid proof`
* _unknown_ &#8658; fail with `need a proof between current and h`
* _true_ &#8658; set _T_ = _T_ &#8746; _(H<sub>h </sub>,C<sub>h </sub>)_

Define _max(T)_ as _max(h, where H<sub>h</sub>_ &#8712; _T)_. For any _T_ with _max(T) = h-1_, there must exist some _X<sub>h </sub>|<sub> </sub>U<sub>h</sub>_ so that _max(update(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>)) = h_.
By induction, there must exist a set of proofs, such that _max(update…(T,...)) = h+n_ for any n.

Bisection can be used to discover this set of proofs. That is, given _max(T) = n_ and _valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>) = unknown_, we then try _update(T, X<sub>b </sub>|<sub> </sub>U<sub>b </sub>)_, where _b = (h+n)/2_. The base case is where _valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>) = true_ and is guaranteed to exist if _h=max(T)+1_.

#### 2.3.3 Closing a Connection

IBC implementations may optionally include the ability to close an IBC connection and prevent further header updates, simply causing _update(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>)_ as defined above to always return _false_.

Closing a connection may break application invariants (such as fungiblity - token vouchers on chain _B_ will no longer be redeemable for tokens on chain _A_) and should only be undertaken in extreme circumstances such as Byzantine behavior of the connected chain.

Closure may be permissioned to an on-chain governance system, an identifiable party on the other chain (such as a signer quorum, although this will not work in some Byzantine cases), or any user who submits an application-specific fraud proof. When a connection is closed, application-specific measures may be undertaken to recover assets held on a Byzantine chain. We defer further discussion to [Appendix D](appendices.md#appendix-d-byzantine-recovery-strategies).
