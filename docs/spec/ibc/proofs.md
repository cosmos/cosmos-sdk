## 2 Proofs

([Back to table of contents](specification.md#contents))

The basis of IBC is the ability to perform efficient proofs of a message packet on-chain and deterministically. All transactions must be attributable and provable without depending on any information outside of the blockchain. We define the following variables: _H<sub>h</sub> _is the signed header at height _h_, _C<sub>h</sub>_ are the consensus rules at height _h_, and _P_ is the unbonding period of this blockchain. _V<sub>k,h</sub>_ is the value stored under key _k_ at height _h_. Note that of all these, only _H<sub>h</sub>_ defines a signature and is thus attributable.

To support an IBC connection, two actors must be able to make the following proofs to each other:

*   given a trusted _H<sub>h</sub>_ and _C<sub>h</sub>_ and an attributable update message _U<sub>h'</sub>_ it is possible to prove _H<sub>h'</sub>_ where _C<sub>h'</sub> = C<sub>h</sub>_ and

<p id="gdcalert1" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert2">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_(now, H<sub>h</sub>) < P_
*   given a trusted _H<sub>h</sub>_ and _C<sub>h</sub>_ and an attributable change message _X<sub>h'</sub>_ it is possible to prove _H<sub>h'</sub>_ where _C<sub>h'</sub> _

<p id="gdcalert2" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert3">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_C<sub>h</sub>_ and

<p id="gdcalert3" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert4">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_(now, H<sub>h</sub>) < P_
*   given a trusted _H<sub>h</sub>_ and a merkle proof _M<sub>k,v,h</sub>_ it is possible to prove _V<sub>k,h</sub>_

It is possible to make use of the structure of BFT consensus to construct extremely lightweight and provable messages _U<sub>h'</sub>_ and _X<sub>h'</sub>_. The implementation of these requirements with Tendermint is defined in Appendix E. Another engine able to provide equally strong guarantees (such as Casper) should be theoretically compatible with IBC, and must define its own set of update/change messages.

The merkle proof _M<sub>k,v,h</sub>_ is a well-defined concept in the blockchain space, and provides a compact proof that the key value pair (_k, v)_ is consistent with a merkle root stored in _H<sub>h</sub>_. Handling the case where _k_ is not in the store requires a separate proof of non-existence, which is not supported by all merkle stores. Thus, we define the proof only as a proof of existence. There is no valid proof for missing keys, and we design the algorithm to work without it.

_valid(H<sub>h </sub>,M<sub>k,v,h </sub>) _

<p id="gdcalert4" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert5">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ [true | false]_

### 2.1   Establishing a Root of Trust

As mentioned in the definitions, all proofs are based on an original assumption. In this case it is _H<sub>h</sub>_ and _C<sub>h</sub>_ for some _h_, where

<p id="gdcalert5" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert6">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_(now, H<sub>h</sub>) < P_.

Any header may be from a malicious chain (eg. shadowing a real chain id with a fake validator set), so a subjective decision is required before establishing a connection. This should be performed by on-chain governance to avoid an exploitable position of trust. Establishing a bidirectional root of trust between two blockchains (A trusts B and B trusts A) is a necessary and sufficient prerequisite for all other IBC activity.

Development of a fully open and decentralized PKI for tracking blockchains is an open research question for future iterations of the IBC protocol.

### 2.2   Following Block Headers

We define two messages _U<sub>h</sub>_ and _X<sub>h</sub>_, which together allow us to securely advance our trust from some known _H<sub>n</sub>_ to a future _H<sub>h</sub>_ where _h > n_. Some implementations may provide the additional limitation that _h = n + 1_, which requires us to process every header. Tendermint allows us to exploit knowledge of the BFT algorithm to only require the additional limitation

<p id="gdcalert6" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert7">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

<sub><em>vals(Cn, Ch ) < ⅓</em></sub>, that each step must have a change of less than one-third of the validator set[[4](./footnotes.md#4)].

Any of these requirements allows us to support IBC for the given block chain. However, by supporting proofs where  _h_-_n > 1_, we can follow the block headers much more efficiently in situations where the majority of blocks do not include an IBC message between chains A and B, and enable low-bandwidth connections to be implemented at very low cost. If there are messages to relay every block, then these collapse to the same case, relaying every header.

Since these messages _U<sub>h</sub>_ and _X<sub>h</sub>_ provide all knowledge of the remote blockchain, we require that they not just be provable, but also attributable. As such any attempt to violate the finality guarantees or provide fake proof can be submitted to the remote blockchain for punishment, in the same manner that any violation of the internal consensus algorithm is punished. This incentive enhances the security guarantees and avoids the nothing-at-stake issue in IBC as well.

More formally, given existing set of trust _T_ = _{(H<sub>i </sub>, C<sub>i </sub>), (H<sub>j </sub>, C<sub>j </sub>), …}_, we must provide:

_valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>) _

<p id="gdcalert7" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert8">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ [true | false | unknown]_

_if H<sub>h-1</sub> _

<p id="gdcalert8" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert9">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_T then _

_valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>) _

<p id="gdcalert9" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert10">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ [true | false]_

_there must exist some U<sub>h</sub> or X<sub>h</sub> that evaluates to true_

_if C<sub>h</sub> _

<p id="gdcalert10" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert11">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_T then_

_   valid(T, U<sub>h </sub>) _

<p id="gdcalert11" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert12">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ false_

and can process update transactions as follows:

_update(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>) _

<p id="gdcalert12" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert13">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ match valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>)_

_   false _

<p id="gdcalert13" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert14">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ return Error("invalid proof")_

_   unknown _

<p id="gdcalert14" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert15">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ return Error("need a proof between current and h")_

_   true _

<p id="gdcalert15" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert16">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_ T _

<p id="gdcalert16" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert17">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_(H<sub>h </sub>,C<sub>h </sub>)_

We define _max(T)_ as _max(h, where H<sub>h</sub>_

<p id="gdcalert17" ><span style="color: red; font-weight: bold">>>>>>  gd2md-html alert: equation: use MathJax/LaTeX if your publishing platform supports it. </span><br>(<a href="#">Back to top</a>)(<a href="#gdcalert18">Next alert</a>)<br><span style="color: red; font-weight: bold">>>>>> </span></p>

_T) _for any _T_ with _max(T) = h-1_. And from above, there must exist some _X<sub>h </sub>|<sub> </sub>U<sub>h</sub>_ so that _max(update(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>)) = h_. By induction, we can see there must exist a set of proofs, such that _max(update…(T,...)) = h+n_ for any n.

We also can see the validity of using bisection as an optimization to discover this set of proofs. That is, given _max(T) = n_ and _valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>) = unknown_, we then try _update(T, X<sub>b </sub>|<sub> </sub>U<sub>b </sub>)_, where _b = (h+n)/2_. The base case is where _valid(T, X<sub>h </sub>|<sub> </sub>U<sub>h </sub>) = true_ and is guaranteed to exist if _h=max(T)+1_.
