------------------------- MODULE counterexample -------------------------

EXTENDS transfer_instance

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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
      receiver |-> "a2",
      sender |-> ""],
  destChannel |-> "buckybucky",
  destPort |-> "buckybucky",
  sourceChannel |-> "zarkozarko",
  sourcePort |-> "buckybucky"]

(* Transition 2 to State3 *)

State3 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "buckybucky",
      denom |-> "atom",
      port |-> "buckybucky"]
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |->
        [channel |-> "buckybucky", denom |-> "atom", port |-> "buckybucky"],
      receiver |-> "a1",
      sender |-> "a2"],
  destChannel |-> "buckybucky",
  destPort |-> "buckybucky",
  sourceChannel |-> "buckybucky",
  sourcePort |-> "buckybucky"]

(* Transition 6 to State4 *)

State4 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "buckybucky",
      denom |-> "atom",
      port |-> "buckybucky"]
  >>
    :> 1
/\ count = 2
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> FALSE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "buckybucky",
                  denom |-> "atom",
                  port |-> "buckybucky"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "buckybucky",
          sourcePort |-> "buckybucky"]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |->
        [channel |-> "zarkozarko", denom |-> "", port |-> "zarkozarko"],
      receiver |-> "a1",
      sender |-> "a1"],
  destChannel |-> "zarkozarko",
  destPort |-> "zarkozarko",
  sourceChannel |-> "",
  sourcePort |-> ""]

(* Transition 7 to State5 *)

State5 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "buckybucky",
      denom |-> "atom",
      port |-> "buckybucky"]
  >>
    :> 1
/\ count = 3
/\ error = TRUE
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> FALSE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "buckybucky",
                  denom |-> "atom",
                  port |-> "buckybucky"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "buckybucky",
          sourcePort |-> "buckybucky"]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> TRUE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "zarkozarko", denom |-> "", port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarkozarko",
          destPort |-> "zarkozarko",
          sourceChannel |-> "",
          sourcePort |-> ""]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |->
        [channel |-> "buckybucky", denom |-> "atom", port |-> "buckybucky"],
      receiver |-> "a1",
      sender |-> ""],
  destChannel |-> "zarkozarko",
  destPort |-> "zarkozarko",
  sourceChannel |-> "",
  sourcePort |-> ""]

(* Transition 7 to State6 *)

State6 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "buckybucky",
      denom |-> "atom",
      port |-> "buckybucky"]
  >>
    :> 1
/\ count = 4
/\ error = TRUE
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> FALSE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "buckybucky",
                  denom |-> "atom",
                  port |-> "buckybucky"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "buckybucky",
          sourcePort |-> "buckybucky"]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> TRUE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "zarkozarko", denom |-> "", port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarkozarko",
          destPort |-> "zarkozarko",
          sourceChannel |-> "",
          sourcePort |-> ""]]
  @@ 4
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> TRUE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "buckybucky",
                  denom |-> "atom",
                  port |-> "buckybucky"],
              receiver |-> "a1",
              sender |-> ""],
          destChannel |-> "zarkozarko",
          destPort |-> "zarkozarko",
          sourceChannel |-> "",
          sourcePort |-> ""]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |->
        [channel |-> "zarkozarko", denom |-> "", port |-> "zarkozarko"],
      receiver |-> "a1",
      sender |-> "a1"],
  destChannel |-> "zarkozarko",
  destPort |-> "zarkozarko",
  sourceChannel |-> "buckybucky",
  sourcePort |-> "zarkozarko"]

(* Transition 0 to State7 *)

State7 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "buckybucky",
      denom |-> "atom",
      port |-> "buckybucky"]
  >>
    :> 1
/\ count = 5
/\ error = TRUE
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "buckybucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> FALSE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "buckybucky",
                  denom |-> "atom",
                  port |-> "buckybucky"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "buckybucky",
          sourcePort |-> "buckybucky"]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> TRUE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "zarkozarko", denom |-> "", port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarkozarko",
          destPort |-> "zarkozarko",
          sourceChannel |-> "",
          sourcePort |-> ""]]
  @@ 4
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> TRUE,
      handler |-> "OnRecvAcknowledgementResult",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "buckybucky",
                  denom |-> "atom",
                  port |-> "buckybucky"],
              receiver |-> "a1",
              sender |-> ""],
          destChannel |-> "zarkozarko",
          destPort |-> "zarkozarko",
          sourceChannel |-> "",
          sourcePort |-> ""]]
  @@ 5
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |->
                "buckybucky",
              denom |-> "atom",
              port |-> "buckybucky"]
          >>
            :> 1,
      error |-> TRUE,
      handler |-> "SendTransfer",
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "zarkozarko", denom |-> "", port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarkozarko",
          destPort |-> "zarkozarko",
          sourceChannel |-> "buckybucky",
          sourcePort |-> "zarkozarko"]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |-> [channel |-> "", denom |-> "", port |-> ""],
      receiver |-> "",
      sender |-> ""],
  destChannel |-> "",
  destPort |-> "",
  sourceChannel |-> "",
  sourcePort |-> ""]

(* The following formula holds true in the last state and violates the invariant *)

InvariantViolation ==
  count >= 5
    /\ BMC!Skolem((\E s1$2 \in DOMAIN history:
      BMC!Skolem((\E s2$2 \in DOMAIN history:
        ~(history[s1$2]["handler"] = history[s2$2]["handler"])))))

================================================================================
\* Created by Apalache on Fri Nov 27 14:00:04 CET 2020
\* https://github.com/informalsystems/apalache
