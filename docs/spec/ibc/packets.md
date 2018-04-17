## 3 Packets

([Back to table of contents](README.md#contents))

IBC uses an asynchronous message passing model that makes no assumptions about network synchrony. Chain _A_ and chain _B_ confirm new blocks independently, and IBC packets from one chain to the other may be delayed or censored arbitrarily. The speed of the IBC packet queue is limited only by the speed of the underlying chains.

The IBC protocol as defined here is payload-agnostic. The packet receiver on chain _B_ decides how to act upon the incoming message, and may add its own application logic to determine which state transactions to apply (or not). Both chains must only agree that the packet has been received and either accepted or rejected, which is determined independently of any application logic.

To facilitate building useful application logic, we introduce a reliable messaging queue (hereafter just referred to as a queue) to allow us to guarantee a cross-chain causal ordering[[5](./references.md#5)] of IBC packets. Causal ordering means that if packet _x_ is processed before packet _y_ on chain _A_, packet _x_ must also be processed before packet _y_ on chain _B_. IBC implements a [vector clock](https://en.wikipedia.org/wiki/Vector_clock) for the restricted case of two processes (in our case, blockchains). 

Formally, given _x_ &#8594; _y_ means _x_ is causally before _y_, and chains A and B, and _a_ &#8658; _b_ means _a_ implies _b_:

_A:send(msg<sub>i </sub>)_ &#8594; _B:receive(msg<sub>i </sub>)_

_B:receive(msg<sub>i </sub>)_ &#8594; _A:receipt(msg<sub>i </sub>)_

_A:send(msg<sub>i </sub>)_ &#8594; _A:send(msg<sub>i+1 </sub>)_

_x_ &#8594; _A:send(msg<sub>i </sub>)_ &#8658;
_x_ &#8594; _B:receive(msg<sub>i </sub>)_

_y_ &#8594; _B:receive(msg<sub>i </sub>)_ &#8658;
_y_ &#8594; _A:receipt(msg<sub>i </sub>)_

Every transaction on the same chain already has a well-defined causality relation (order in history). IBC provides an ordering guarantee across two chains which can be used to reason about the combined state of both chains as a whole.

For example, an application may wish to allow a single fungible asset to be transferred between and held on multiple blockchains while preserving conservation of supply. The application can mint asset vouchers on chain _B_ when a particular IBC packet is committed to chain _B_, and require outgoing sends of that packet on chain _A_ to escrow an equal amount of the asset on chain _A_ until the vouchers are later redeemed back to chain _A_ with an IBC packet in the reverse direction. This ordering guarantee along with correct application logic can ensure that total supply is preserved across both chains and that any vouchers minted on chain _B_ can later be redeemed back to chain _A_.

This section provides a high-level specification of the queue interface and a list of the necessary proofs. To implement wire-compatible IBC, chain _A_ and chain _B_ must also use a common encoding format. An example binary encoding format can be found in Appendix C.

### 3.1 Definitions

We introduce the abstraction of an IBC _connection_: a set of the required components to facilitate bidirectional communication between two blockchains _A_ and _B_.

An IBC connection consists of four distinct queues, two on each chain:

_Outgoing<sub>A</sub>_: Outgoing IBC packets from chain _A_ to chain _B_, stored on chain _A_

_Incoming<sub>A</sub>_: Execution logs for incoming IBC packets from chain _B_, stored on chain _A_

_Outgoing<sub>B</sub>_: Outgoing IBC packets from chain _B_ to chain _A_, stored on chain _B_

_Incoming<sub>B</sub>_: Execution logs for incoming IBC packets from chain _A_, stored on chain _B_

### 3.2 Requirements

A queue can be conceptualized as a slice of an infinite array. Two numerical indices - _q<sub>head</sub>_ and _q<sub>tail</sub>_ - bound the slice, such that for every _index_ where _head <= index < tail_, there is a queue element _q[q<sub>index</sub>]_. Elements can be appended to the tail (end) and removed from the head (beginning). We introduce one further method, _advance_, to facilitate efficient queue cleanup.

Each IBC-supporting blockchain must implement a reliable ordered packet queue with the following interface specification:

**init**  
> set _q<sub>head</sub>_ = _0_  
> set _q<sub>tail</sub>_ = _0_

**peek** &#8658; **e** 
> match _q<sub>head</sub> == q<sub>tail</sub>_ with  
>    _true_ &#8658; return _nil_  
>   _false_ &#8658;  return _q[q<sub>head</sub>]_

**pop** &#8658; **e** 
> match _q<sub>head</sub> == q<sub>tail</sub>_ with   
>   _true_ &#8658; return _nil_  
>   _false_ &#8658; set _q<sub>head</sub>_ = _q<sub>head</sub> + 1_; return _q[q<sub>head</sub>-1]_

**retrieve(i)** &#8658; **e**
> match _q<sub>head</sub> <= i < q<sub>tail</sub>_ with   
>   _true_  &#8658; return _q<sub>i</sub>_  
>   _false_  &#8658; return _nil_

**push(e)**
> set _q[q<sub>tail</sub>]_ = _e_; set _q<sub>tail</sub>_ = _q<sub>tail</sub> + 1_

**advance(i)**
> set _q<sub>head</sub>_ = _i_; set _q<sub>tail</sub>_ = _max(q<sub>tail</sub>, i)_

**head** &#8658; **i**
> return _q<sub>head</sub>_

**tail** &#8658; **i**
> return _q<sub>tail</sub>_

In order to provide the ordering guarantees specified above, each blockchain utilizing the IBC protocol must provide proofs that particular IBC packets have been stored at particular indices in the outgoing packet queue, and particular IBC packet execution results have been stored at particular indices in the incoming packet queue.

We use the previously-defined Merkle proof _M<sub>k,v,h</sub>_ to provide the requisite proofs. In order to do so, we must define a unique, deterministic key in the Merkle store for each message in the queue:

**key**: _(queue name, [head|tail|index])_

The index is stored as a fixed-length unsigned integer in big endian format, so that the lexicographical order of the byte representation of the key is consistent with their sequence number. This allows us to quickly iterate over the queue, as well as prove the content of a packet (or lack of packet) at a given sequence. _head_ and _tail_ are two special constants that store an integer index, and are chosen such that their serializated representation cannot collide with that of any possible index.

Once written to the queue, a packet must be immutable (except for deletion when popped from the queue). That is, if a value _v_ is written to a queue, then every valid proof _M<sub>k,v,h </sub>_ must refer to the same _v_. In practice, this means that an IBC implementation must ensure that only the IBC module can write to the IBC subspace of the blockchain's Merkle store. This property is essential to safely process asynchronous messages.

Each incoming & outgoing queue must be provably associated with another uniquely identified chain, so that an observer can prove that a message was intended for that chain and only that chain. This can easily be done by prefixing the queue keys in the Merkle store with a string unique to the other chain, such as the chain identifier or the hash of the genesis block.

These two queues have different purposes and store elements of different types. By parsing the key of a Merkle proof, a recipient can uniquely identify which queue, if any, this message belongs to. We now define _k =_ _(remote id, [send|receipt], index)_. This tuple is used to route and verify every message, before the contents of the packet are processed by the appropriate application logic.

We define every message in a _send queue_ to consist of two fields: an enumerable _type_, and an opaque _payload_. The IBC protocol relies on the type for routing, and lets the appropriate module process the data as it sees fit. The _receipt queue_ stores if it was an error, an optional error code, and an optional return value. We use the same index as the received message, so that the results of _A:q<sub>B.send</sub>[i]_ are stored at _B:q<sub>A.receipt</sub>[i]_. (read: the message at index _i_ in the _send_ queue for chain B as stored on chain A)

_V<sub>send</sub> = (type, data)_

_V<sub>receipt</sub> = (result, [success|error code])_

### 3.3 Sending a packet

{ todo: cleanup wording }

A proper implementation of IBC requires all relevant state to be encapsulated, so that other modules can only interact with it via a fixed API (to be defined in the next sections) rather than directly mutating internal state. This allows the IBC module to provide security guarantees.

Sending an IBC packet involves an application module calling the send method of the IBC module with a packet and a destination chain id.  The IBC module must ensure that the destination chain was already properly registered, and that the calling module has permission to write this packet. If so, the IBC module simply pushes the packet to the tail of the _send queue_, which enables all the proofs described above.

The permissioning of which module can write which packet can be defined per type, so this module can maintain any application-level invariants related to this area. Thus, the "coin" module can maintain the constant supply of tokens, while another module can maintain its own invariants, without IBC messages providing a means to escape their encapsulations. The IBC module must associate every supported message type with a particular handler (_f<sub>type</sub>_) and return an error for unsupported types.

_(IBCsend(D, type, data)_ &#8658; _Success)_
  &#8658; _push(q<sub>D.send</sub> ,V<sub>send</sub>{type, data})_

We also consider how a given blockchain _A_ is expected to receive the packet from a source chain _S_ with a merkle proof, given the current set of trusted headers for that chain, _T<sub>S</sub>_:

_A:IBCreceive(S, M<sub>k,v,h</sub>)_ &#8658; _match_
  * _q<sub>S.receipt</sub> =_ &#8709; &#8658; _Error("unregistered sender"),_
  * _k = (\_, reciept, \_)_ &#8658; _Error("must be a send"),_
  * _k = (d, \_, \_) and d_ &#8800; _A_ &#8658; _Error("sent to a different chain"),_
  * _k = (\_, send, i) and head(q<sub>S.receipt</sub>)_ &#8800; _i_ &#8658; _Error("out of order"),_
  * _H<sub>h</sub>_ &#8713; _T<sub>S</sub>_ &#8658; _Error("must submit header for height h"),_
  * _valid(H<sub>h</sub> ,M<sub>k,v,h </sub>) = false_ &#8658; _Error("invalid merkle proof"),_
  * _v = (type, data)_ &#8658; _(result, err) := f<sub>type</sub>(data); push(q<sub>S.receipt </sub>, (result, err)); Success_

Note that this requires not only an valid proof, but also that the proper header as well as all prior messages were previously submitted. This returns success upon accepting a proper message, even if the message execution returned an error (which must then be relayed to the sender).

### 3.4 Receiving a packet

{ todo: cleanup logic }

When we wish to create a transaction that atomically commits or rolls back across two chains, we must look at the receipts from sending the original message. For example, if I want to send tokens from Alice on chain A to Bob on chain B, chain A must decrement Alice's account _if and only if_ Bob's account was incremented on chain B. We can achieve that by storing a protected intermediate state on chain A, which is then committed or rolled back based on the result of executing the transaction on chain B.

To do this requires that we not only provable send a message from chain A to chain B, but provably return the result of that message (the receipt) from chain B to chain A. As one noticed above in the implementation of _IBCreceive_, if the valid IBC message was sent from A to B, then the result of executing it, even if it was an error, is stored in _B:q<sub>A.receipt</sub>_. Since the receipts are stored in a queue with the same key construction as the sending queue, we can generate the same set of proofs for them, and perform a similar sequence of steps to handle a receipt coming back to _S_ for a message previously sent to _A_:

_S:IBCreceipt(A, M<sub>k,v,h</sub>)_ &#8658; _match_
  * _q<sub>A.send</sub> =_ &#8709; &#8658; _Error("unregistered sender"),_
  * _k = (\_, send, \_)_  &#8658; _Error("must be a recipient"),_
  * _k = (d, \_, \_) and d_ &#8800; _S_ &#8658; _Error("sent to a different chain"),_
  * _H<sub>h</sub>_ &#8713; _T<sub>A</sub>_ &#8658; _Error("must submit header for height h"),_
  * _not valid(H<sub>h </sub>, M<sub>k,v,h </sub>)_ &#8658; _Error("invalid merkle proof"),_
  * _k = (\_, receipt, head|tail)_ &#8658; _Error("only accepts message proofs"),_
  * _k = (\_, receipt, i) and head(q<sub>S.send</sub>)_ &#8800; _i_ &#8658; _Error("out of order"),_
  * _v = (\_, error)_ &#8658; _(type, data) := pop(q<sub>S.send </sub>); rollback<sub>type</sub>(data); Success_
  * _v = (res, success)_ &#8658; _(type, data) := pop(q<sub>S.send </sub>); commit<sub>type</sub>(data, res); Success_

This enforces that the receipts are processed in order, to allow some the application to make use of some basic assumptions about ordering. It also removes the message from the send queue, as there is now proof it was processed on the receiving chain and there is no more need to store this information.

![Successful Transaction](images/Receipts.png)

![Rejected Transaction](images/ReceiptError.png)

### 3.5 Packet relayer

{ todo: cleanup wording }

The blockchain itself only records the _intention_ to send the given message to the recipient chain, it doesn't make any network connections as that would add unbounded delays and non-determinism into the state machine. We define the concept of a _relay_ process that connects two chain by querying one for all proofs needed to prove outgoing messages and submit these proofs to the recipient chain.

The relay process must have access to accounts on both chains with sufficient balance to pay for transaction fees but needs no other permissions. Many _relay_ processes may run in parallel without violating any safety consideration. However, they will consume unnecessary fees if they submit the same proof multiple times, so some minimal coordination is ideal.

As an example, here is a naive algorithm for relaying send messages from A to B, without error handling. We must also concurrently run the relay of receipts from B back to A, in order to complete the cycle. Note that all reads of variables belonging to a chain imply queries and all function calls imply submitting a transaction to the blockchain.

```
while true
   pending := tail(A:q<sub>B.send</sub>)
   received := tail(B:q<sub>A.receive</sub>)
   if pending > received
       U<sub>h</sub> := A:latestHeader
       B:updateHeader(U<sub>h</sub>)
       for i :=received...pending
           k := (B, send, i)
           packet := A:M<sub>k,v,h</sub>
           B:IBCreceive(A, packet)
   sleep(desiredLatency)
```

Note that updating a header is a costly transaction compared to posting a merkle proof for a known header. Thus, a process could wait until many messages are pending, then submit one header along with multiple merkle proofs, rather than a separate header for each message. This decreases total computation cost (and fees) at the price of additional latency and is a trade-off each relay can dynamically adjust.

In the presence of multiple concurrent relays, any given relay can perform local optimizations to minimize the number of headers it submits, but remember the frequency of header submissions defines the latency of the packet transfer.

Indeed, it is ideal if each user that initiates the creation of an IBC packet also relays it to the recipient chain. The only constraint is that the relay must be able to pay the appropriate fees on the destination chain. However, in order to avoid bottlenecks, a group may sponsor an account to pay fees for a public relayer that moves all unrelayed packets (perhaps with a high latency).
