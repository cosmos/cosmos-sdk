## 3 Messaging Queue

([Back to table of contents](specification.md#contents))

Messaging in distributed systems is a deeply researched field and a primitive upon which many other systems are built. We can model asynchronous message passing, and make no timing assumptions on the communication channels. By doing this, we allow each zone to move at its own speed, unblocked by any other zone, but able to communicate as fast as the network allows at that moment.

Another benefit of using message passing as our primitive, is that the receiver decides how to act upon the incoming message. Just because one zone sends a message and we have an IBC connection with this zone, doesn't mean we have to execute the requested action. Each zone can add its own business logic upon receiving the message to decide whether to accept or reject the message. To maintain consistency, both sides must only agree on the proper state transitions associated with accepting or rejecting.

This encapsulation is very difficult to impossible to achieve in a shared-state scenario. Message passing allows each zone to ensure its security and autonomy, while simultaneously allowing the different systems to work as one whole. This can be seen as an analogue to a microservices architecture, but across organizational boundaries.

To build useful algorithms upon a provable asynchronous messaging primitive, we introduce a reliable messaging queue (hereafter just referred to as a queue), typical in asynchronous message passing, to allow us to guarantee a causal ordering[[5](./footnotes.md#5)], and avoid blocking.

Causal ordering means that if _x_ is causally before _y_ on chain A, it must also be on chain B. Many events may happen concurrently (unrelated tx on two different blockchains) with no causality relation, but every transaction on the same chain has a clear causality relation (same as the order in the blockchain).

Message passing implies a causal ordering over multiple chains and these can be important for reasoning on the system. Given _x _

<p id="gdcalert18" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert19">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ y_ means _x_ is causally before _y_, and chains A and B, and _a_

<p id="gdcalert19" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert20">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_b_ means _a_ implies _b_:

_A:send(msg<sub>i </sub>)_

<p id="gdcalert20" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert21">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ B:receive(msg<sub>i </sub>)_

_B:receive(msg<sub>i </sub>)_

<p id="gdcalert21" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert22">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ A:receipt(msg<sub>i </sub>)_

_A:send(msg<sub>i </sub>)_

<p id="gdcalert22" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert23">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_A:send(msg<sub>i+1 </sub>)_

_x_

<p id="gdcalert23" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert24">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_A:send(msg<sub>i </sub>)_

<p id="gdcalert24" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert25">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

 _x_

<p id="gdcalert25" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert26">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_B:receive(msg<sub>i </sub>)_

_y_

<p id="gdcalert26" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert27">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_B:receive(msg<sub>i </sub>)_

<p id="gdcalert27" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert28">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

 _y_

<p id="gdcalert28" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert29">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_A:receipt(msg<sub>i </sub>)_



<p id="gdcalert29" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/Cosmos-IBC0.png). Store image on your image server and adjust path/filename if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert30">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/Cosmos-IBC0.png "image_tooltip")


([https://en.wikipedia.org/wiki/Vector_clock](https://en.wikipedia.org/wiki/Vector_clock))

In this section, we define an efficient implementation of a secure, reliable messaging queue.

### 3.1   Merkle Proofs for Queues

Given the three proofs we have available, we make use of the most flexible one, _M<sub>k,v,h</sub>_, to provide proofs for a message queue. To do so, we must define a unique, deterministic, and predictable key in the merkle store for each message in the queue. We also define a clearly defined format for the content of each message in the queue, which can be parsed by all chains participating in IBC. The key format and queue ordering are conceptually explained here. The binary encoding format can be found in Appendix C.

We can visualize a queue as a slice pointing into an infinite sized array. It maintains a head and a tail pointing to two indexes, such that there is data for every index where _head <= index < tail_. Data is pushed to the tail and popped from the head. Another method, _advance_, is introduced to pop all messages until _i_, and is useful for cleanup:

**init**: _q<sub>head</sub> = q<sub>tail</sub> = 0_

**peek **

<p id="gdcalert30" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert31">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

**m**: _if q<sub>head</sub> = q<sub>tail</sub> { return None } else { return q[q<sub>head</sub>] }_

**pop **

<p id="gdcalert31" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert32">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

**m**: _if q<sub>head</sub> = q<sub>tail</sub> { return None } else { q<sub>head</sub>++; return q[q<sub>head</sub>-1] }_

**push(m)**: _q[q<sub>tail</sub>] = m; q<sub>tail</sub>++_

**advance(i)**:  _q<sub>head</sub> = i; q<sub>tail</sub> = max(q<sub>tail </sub>, i)_

**head **

<p id="gdcalert32" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert33">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

**i**: _q<sub>head</sub>_

**tail**

<p id="gdcalert33" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert34">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

**i**: _q<sub>tail</sub>_

Based upon this needed functionality, we define a set of keys to be stored in the merkle tree, which allows us to efficiently implement and prove any of the above queries.

**Key:_ (queue name, [head|tail|index])_**

The index is stored as a fixed-length unsigned integer in big endian format, so that the lexicographical order of the byte representation of the key is consistent with their sequence number. This allows us to quickly iterate over the queue, as well as prove the content of a packet (or lack of packet) at a given sequence. _head_ and _tail_ are two special constants that store an integer index, and are chosen such that their serialization cannot collide with any possible index.

A message queue is simply a set of serialized packets stored at predefined keys in a merkle store, which can produce proofs for any key. Once a packet is written it must be immutable (except for deleting when popped from the queue). That is, if a value _v_ is written to a queue, then every valid proof _M<sub>k,v,h </sub>_must refer to the same _v_. This property is essential to safely process asynchronous messages.

Every IBC implementation must provide a protected subspace of the merkle store for use by each queue that cannot be affected by other modules.

### 3.2   Naming Queues

As mentioned above, in order for the receiver to unambiguously interpret the merkle proofs, we need a unique, deterministic, and predictable key in the merkle store for each message in the queue. We explained how the indexes are generated to provide each message in a queue a unique key, and mentioned the need for a unique name for each queue.

The queue name must be unambiguously associated with a given connection to another chain, so an observer can prove if a message was intended for chain A or chain B. In order to do so, upon registration of a connection with a remote chain, we create two queues with different names (prefixes).

*   _ibc:<chain id of A>:send_ - all outgoing packets destined to chain A
*   _ibc:<chain id of A>:receipt_ - the results of executing the packets received from chain A

These two queues have different purposes and store messages of different types. By parsing the key of a merkle proof, a recipient can uniquely identify which queue, if any, this message belongs to. We now define _k =_ (_remote id, [send|receipt], index)_. This tuple is used to route and verify every message, before the contents of the packet are processed by the appropriate application logic.

### 3.3   Message Contents

Up to this point, we have focused on the semantics of the message key, and how we can produce a unique identifier for every possible message in every possible connection. The actual data written at the location has been left as an opaque blob, put by providing some structure to the messages, we can enable more functionality.

We define every message in a _send queue _to consist of a well-known type and opaque data. The IBC protocol relies on the type for routing, and lets the appropriate module process the data as it sees fit. The _receipt queue_ stores if it was an error, an optional error code, and an optional return value. We use the same index as the received message, so that the results of _A:q<sub>B.send</sub>[i]_ are stored at _B:q<sub>A.receipt</sub>[i]_. (read: the message at index _i_ in the _send_ queue for chain B as stored on chain A)

_V<sub>send</sub> = (type, data)_

_V<sub>receipt</sub> = (result, [success|error code])_

### 3.4   Sending a Message

A proper implementation of IBC requires all relevant state to be encapsulated, so that other modules can only interact with it via a fixed API (to be defined in the next sections) rather than directly mutating internal state. This allows the IBC module to provide security guarantees.

Sending an IBC packet involves an application module calling the send method of the IBC module with a packet and a destination chain id.  The IBC module must ensure that the destination chain was already properly registered, and that the calling module has permission to write this packet. If so, the IBC module simply pushes the packet to the tail of the _send_ _queue_, which enables all the proofs described above.

The permissioning of which module can write which packet can be defined per type, so this module can maintain any application-level invariants related to this area. Thus, the "coin" module can maintain the constant supply of tokens, while another module can maintain its own invariants, without IBC messages providing a means to escape their encapsulations. The IBC module must associate every supported message type with a particular handler (_f<sub>type</sub>_) and return an error for unsupported types.

_(IBCsend(D, type, data) _

<p id="gdcalert34" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert35">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ Success)_


<p id="gdcalert35" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert36">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ push(q<sub>D.send</sub> ,V<sub>send</sub>{type, data})_

We also consider how a given blockchain _A _is expected to receive the packet from a source chain _S_ with a merkle proof, given the current set of trusted headers for that chain, _T<sub>S</sub>_:

_A:IBCreceive(S, M<sub>k,v,h</sub>) _

<p id="gdcalert36" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert37">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ match_

_q<sub>S.receipt</sub> =_

<p id="gdcalert37" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert38">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>



<p id="gdcalert38" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert39">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("unregistered sender"), _

_   k = (_, reciept, _) _

<p id="gdcalert39" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert40">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("must be a send"),_

_   k = (d, _, _) and d _

<p id="gdcalert40" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert41">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_A_

<p id="gdcalert41" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert42">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("sent to a different chain"),_

_   k = (_, send, i) and head(q<sub>S.receipt</sub>) _

<p id="gdcalert42" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert43">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_i_

<p id="gdcalert43" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert44">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("out of order"),_

_   H<sub>h</sub> _

<p id="gdcalert44" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert45">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_T<sub>S</sub> _

<p id="gdcalert45" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert46">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("must submit header for height h"),_

_   valid(H<sub>h</sub> ,M<sub>k,v,h </sub>) = false _

<p id="gdcalert46" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert47">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("invalid merkle proof"),_

_   v = (type, data) _

<p id="gdcalert47" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert48">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_(result, err) :=f<sub>type</sub>(data); push(q<sub>S.receipt </sub>, (result, err)); Success   _

Note that this requires not only an valid proof, but also that the proper header as well as all prior messages were previously submitted. This returns success upon accepting a proper message, even if the message execution returned an error (which must then be relayed to the sender).

### 3.5   Receipts

When we wish to create a transaction that atomically commits or rolls back across two chains, we must look at the receipts from sending the original message. For example, if I want to send tokens from Alice on chain A to Bob on chain B, chain A must decrement Alice's account _if and only if_ Bob's account was incremented on chain B. We can achieve that by storing a protected intermediate state on chain A, which is then committed or rolled back based on the result of executing the transaction on chain B.

To do this requires that we not only provable send a message from chain A to chain B, but provably return the result of that message (the receipt) from chain B to chain A. As one noticed above in the implementation of _IBCreceive_, if the valid IBC message was sent from A to B, then the result of executing it, even if it was an error, is stored in _B:q<sub>A.receipt</sub>_. Since the receipts are stored in a queue with the same key construction as the sending queue, we can generate the same set of proofs for them, and perform a similar sequence of steps to handle a receipt coming back to _S_ for a message previously sent to _A_:

_S:IBCreceipt(A, M<sub>k,v,h</sub>) _

<p id="gdcalert48" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert49">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ match_

_q<sub>A.send</sub> =_

<p id="gdcalert49" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert50">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>



<p id="gdcalert50" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert51">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("unregistered sender"), _

_   k = (_, send, _) _

<p id="gdcalert51" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert52">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("must be a recipient"),_

_   k = (d, _, _) and d _

<p id="gdcalert52" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert53">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_S_

<p id="gdcalert53" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert54">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("sent to a different chain"),_

_H<sub>h</sub> _

<p id="gdcalert54" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert55">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_T<sub>A</sub> _

<p id="gdcalert55" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert56">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("must submit header for height h"),_

_   not valid(H<sub>h</sub> ,M<sub>k,v,h </sub>) _

<p id="gdcalert56" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert57">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("invalid merkle proof"),_

_   k = (_, receipt, head|tail)_

<p id="gdcalert57" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert58">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("only accepts message proofs"),_

_   k = (_, receipt, i) and head(q<sub>S.send</sub>) _

<p id="gdcalert58" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert59">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_i_

<p id="gdcalert59" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert60">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_Error("out of order"),_

_   v = (_, error) _

<p id="gdcalert60" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert61">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_(type, data) := pop(q<sub>S.send </sub>); rollback<sub>type</sub>(data); Success_

_   v = (res, success) _

<p id="gdcalert61" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert62">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_(type, data) := pop(q<sub>S.send </sub>); commit<sub>type</sub>(data, res); Success_

This enforces that the receipts are processed in order, to allow some the application to make use of some basic assumptions about ordering. It also removes the message from the send queue, as there is now proof it was processed on the receiving chain and there is no more need to store this information.



<p id="gdcalert62" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/Cosmos-IBC1.png). Store image on your image server and adjust path/filename if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert63">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/Cosmos-IBC1.png "image_tooltip")




<p id="gdcalert63" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: inline image link here (to images/Cosmos-IBC2.png). Store image on your image server and adjust path/filename if necessary. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert64">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>


![alt_text](images/Cosmos-IBC2.png "image_tooltip")


### 3.6   Relay Process

The blockchain itself only records the _intention_ to send the given message to the recipient chain, it doesn't make any network connections as that would add unbounded delays and non-determinism into the state machine. We define the concept of a _relay_ process that connects two chain by querying one for all proofs needed to prove outgoing messages and submit these proofs to the recipient chain.

The relay process must have access to accounts on both chains with sufficient balance to pay for transaction fees but needs no other permissions. Many _relay_ processes may run in parallel without violating any safety consideration. However, they will consume unnecessary fees if they submit the same proof multiple times, so some minimal coordination is ideal.

As an example, here is a naive algorithm for relaying send messages from A to B, without error handling. We must also concurrently run the relay of receipts from B back to A, in order to complete the cycle. Note that all reads of variables belonging to a chain imply queries and all function calls imply submitting a transaction to the blockchain.

_while true_

_   pending := tail(A:q<sub>B.send</sub>)_

_   received := tail(B:q<sub>A.receive</sub>)_

_   if pending > received_

_       U<sub>h</sub> := A:latestHeader_

_       B:updateHeader(U<sub>h</sub>)_

_       for i :=received...pending_

_           k := (B, send, i)_

_           packet := A:M<sub>k,v,h</sub>_

_           B:IBCreceive(A, packet)_

_   sleep(desiredLatency)_

Note that updating a header is a costly transaction compared to posting a merkle proof for a known header. Thus, a process could wait until many messages are pending, then submit one header along with multiple merkle proofs, rather than a separate header for each message. This decreases total computation cost (and fees) at the price of additional latency and is a trade-off each relay can dynamically adjust.

In the presence of multiple concurrent relays, any given relay can perform local optimizations to minimize the number of headers it submits, but remember the frequency of header submissions defines the latency of the packet transfer.

Indeed, it is ideal if each user that initiates the creation of an IBC packet also relays it to the recipient chain. The only constraint is that the relay must be able to pay the appropriate fees on the destination chain. However, in order to avoid bottlenecks, a group may sponsor an account to pay fees for a public relayer that moves all unrelayed packets (perhaps with a high latency).
