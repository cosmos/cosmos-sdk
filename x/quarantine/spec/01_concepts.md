<!--
order: 1
-->

# Concepts

## Quarantined Account

A quarantined account is one that has elected to not receive funds transfers until the transfer has been accepted.

When funds are sent using the `x/bank` module keeper's `SendCoins` or `InputOutputCoins` functions (e.g. from a `Send` or `MultiSend` Tx),
if the receiver is quarantined, the funds are sent to an intermediary quarantined funds holder account and a record of the quarantined funds is made.
Later, the receiver can `Approve` or `Decline` the funds as they see fit.

## Opt-In

An account becomes quarantined when the account owner issues an `OptIn` Tx.
An account can later opt out by issuing an `OptOut` Tx.
Opting in or out does not affect any previously quarantined funds.

## Quarantined Funds

Quarantined funds can either be accepted or declined (or ignored) by the receiver.

### Accept Funds

When a receiver `Accept`s funds, all fully approved quarantined funds from each sender are transferred to the receiver.
Quarantined funds are fully approved when all senders involved in the transfer of those funds have been part of an `Accept` from the receiver.
Funds quarantined for a receiver are aggregated by sender (or set of senders in the case of a `MultiSend`).
That is, if a sender issues two different transfers to a receiver, they are quarantined together and the receiver only needs to issue a single `Accept` for them.

### Decline Funds

When a receiver `Decline`s funds, all quarantined funds from each sender are marked as declined.
Declined funds remain held by the quarantined fund holder account and can later be accepted.
Declined funds are not returned by the `QuarantinedFunds` query unless the query params included a specific sender.
The decline indicator is reset to `false` if new funds are quarantined (to the same receiver from the same sender) and auto-decline is not set up.

## Auto-Responses

A quarantined account can set up auto-accept from known trusted senders, and auto-decline from known untrusted senders.
An auto-response is unique for a given receiver and sender and only applies in one direction.
That is, the auto-responses that one receiver has defined, do not affect any other accounts.

If funds are sent to a quarantined account from an auto-accept sender, the transfer occurs as if the receiver weren't quarantined.
When there are multiple senders, the funds are quarantined unless the receiver has auto-accept for **ALL** of the senders.

If funds are sent to a quarantined account from an auto-decline sender, the funds are quarantined and marked as declined.
When there are multiple senders, the funds are declined if the receiver has auto-decline for **ANY** of the senders.
