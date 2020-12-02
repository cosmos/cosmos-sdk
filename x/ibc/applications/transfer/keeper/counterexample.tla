------------------------- MODULE counterexample -------------------------

EXTENDS relay_tests

(* Initial state *)

State1 ==
TRUE
(* Transition 0 to State2 *)

State2 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
/\ count = 0
/\ error = FALSE
/\ handler = ""
/\ history = 0
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
      receiver |-> "a2",
      sender |-> "a1"],
  destChannel |-> "channel-1",
  destPort |-> "transfer",
  sourceChannel |-> "channel-0",
  sourcePort |-> "transfer"]

(* Transition 2 to State3 *)

State3 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
      denom |-> "atom",
      port |-> "transfer"]
  >>
    :> 1
/\ count = 1
/\ error = FALSE
/\ handler = "OnRecvPacket"
/\ history = 0
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |->
        [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
      receiver |-> "a1",
      sender |-> "a2"],
  destChannel |-> "channel-0",
  destPort |-> "transfer",
  sourceChannel |-> "channel-1",
  sourcePort |-> "transfer"]

(* Transition 7 to State4 *)

State4 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
      denom |-> "atom",
      port |-> "transfer"]
  >>
    :> 2
/\ count = 2
/\ error = FALSE
/\ handler = "OnTimeoutPacket"
/\ history = 0
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 2,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      error |-> FALSE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "channel-0",
          destPort |-> "transfer",
          sourceChannel |-> "channel-1",
          sourcePort |-> "transfer"]]
/\ p = [data |->
    [amount |-> 2,
      denomTrace |->
        [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
      receiver |-> "a3",
      sender |-> "a2"],
  destChannel |-> "channel-0",
  destPort |-> "transfer",
  sourceChannel |-> "channel-1",
  sourcePort |-> "transfer"]

(* Transition 10 to State5 *)

State5 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
      denom |-> "atom",
      port |-> "transfer"]
  >>
    :> 4
/\ count = 3
/\ error = FALSE
/\ handler = "OnRecvAcknowledgementError"
/\ history = 0
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 2,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      error |-> FALSE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "channel-0",
          destPort |-> "transfer",
          sourceChannel |-> "channel-1",
          sourcePort |-> "transfer"]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 4,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 2,
      error |-> FALSE,
      handler |-> "OnRecvAcknowledgementError",
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a3",
              sender |-> "a2"],
          destChannel |-> "channel-0",
          destPort |-> "transfer",
          sourceChannel |-> "channel-1",
          sourcePort |-> "transfer"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |->
        [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
      receiver |-> "a2",
      sender |-> "a2"],
  destChannel |-> "channel-1",
  destPort |-> "transfer",
  sourceChannel |-> "ethereum-hub",
  sourcePort |-> "bitcoin-hub"]

(* Transition 12 to State6 *)

State6 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
      denom |-> "atom",
      port |-> "transfer"]
  >>
    :> 4
/\ count = 4
/\ error = FALSE
/\ handler = "OnRecvAcknowledgementResult"
/\ history = 0
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 2,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      error |-> FALSE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "channel-0",
          destPort |-> "transfer",
          sourceChannel |-> "channel-1",
          sourcePort |-> "transfer"]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 4,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 2,
      error |-> FALSE,
      handler |-> "OnRecvAcknowledgementError",
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a3",
              sender |-> "a2"],
          destChannel |-> "channel-0",
          destPort |-> "transfer",
          sourceChannel |-> "channel-1",
          sourcePort |-> "transfer"]]
  @@ 4
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 4,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 4,
      error |-> FALSE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "ethereum-hub",
          sourcePort |-> "bitcoin-hub"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |->
        [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
      receiver |-> "a1",
      sender |-> "a2"],
  destChannel |-> "channel-0",
  destPort |-> "transfer",
  sourceChannel |-> "channel-0",
  sourcePort |-> "transfer"]

(* Transition 1 to State7 *)

State7 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
      denom |-> "atom",
      port |-> "transfer"]
  >>
    :> 3
  @@ <<
    [channel |-> "channel-0", id |-> "", port |-> "transfer"], [channel |->
        "channel-1",
      denom |-> "atom",
      port |-> "transfer"]
  >>
    :> 1
/\ count = 5
/\ error = FALSE
/\ handler = "SendTransfer"
/\ history = 0
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> FALSE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 2,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      error |-> FALSE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "channel-0",
          destPort |-> "transfer",
          sourceChannel |-> "channel-1",
          sourcePort |-> "transfer"]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 4,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 2,
      error |-> FALSE,
      handler |-> "OnRecvAcknowledgementError",
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a3",
              sender |-> "a2"],
          destChannel |-> "channel-0",
          destPort |-> "transfer",
          sourceChannel |-> "channel-1",
          sourcePort |-> "transfer"]]
  @@ 4
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 4,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 4,
      error |-> FALSE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "channel-1",
          destPort |-> "transfer",
          sourceChannel |-> "ethereum-hub",
          sourcePort |-> "bitcoin-hub"]]
  @@ 5
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 3
          @@ <<
            [channel |-> "channel-0", id |-> "", port |-> "transfer"], [channel |->
                "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 1,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "channel-1",
              denom |-> "atom",
              port |-> "transfer"]
          >>
            :> 4,
      error |-> FALSE,
      handler |-> "SendTransfer",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "channel-1", denom |-> "atom", port |-> "transfer"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "channel-0",
          destPort |-> "transfer",
          sourceChannel |-> "channel-0",
          sourcePort |-> "transfer"]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |->
        [channel |-> "transfer", denom |-> "eth", port |-> "transfer"],
      receiver |-> "",
      sender |-> "a1"],
  destChannel |-> "bitcoin-hub",
  destPort |-> "channel-0",
  sourceChannel |-> "channel-0",
  sourcePort |-> "channel-0"]

(* The following formula holds true in the last state and violates the invariant *)

InvariantViolation ==
  count >= 5
    /\ (\A s1$2 \in DOMAIN history:
      \A s2$2 \in DOMAIN history:
        s1$2 = s2$2 \/ ~(history[s1$2]["handler"] = history[s2$2]["handler"]))
    /\ (\A s$2 \in DOMAIN history:
      s$2 <= 0
        \/ (history[s$2]["error"] = FALSE
          /\ history[s$2]["packet"]["data"]["amount"] > 0))

================================================================================
\* Created by Apalache on Wed Dec 02 21:29:27 CET 2020
\* https://github.com/informalsystems/apalache
