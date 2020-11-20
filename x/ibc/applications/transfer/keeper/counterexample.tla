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
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |->
        [channel |-> "zarkozarko", denom |-> "atom", port |-> "zarkozarko"],
      receiver |-> "a1",
      sender |-> "a2"],
  destChannel |-> "buckybucky",
  destPort |-> "buckybucky",
  sourceChannel |-> "zarkozarko",
  sourcePort |-> "zarkozarko"]

(* Transition 0 to State3 *)

State3 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
/\ count = 1
/\ error = TRUE
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
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 1
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
      receiver |-> "a1",
      sender |-> ""],
  destChannel |-> "buckybucky",
  destPort |-> "buckybucky",
  sourceChannel |-> "buckybucky",
  sourcePort |-> "buckybucky"]

(* Transition 2 to State4 *)

State4 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
/\ count = 2
/\ error = TRUE
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
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 1
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 2
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
      error |-> TRUE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "buckybucky",
          sourcePort |-> "buckybucky"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
      receiver |-> "a2",
      sender |-> ""],
  destChannel |-> "",
  destPort |-> "",
  sourceChannel |-> "zarkozarko",
  sourcePort |-> "zarkozarko"]

(* Transition 2 to State5 *)

State5 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
/\ count = 3
/\ error = TRUE
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
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 1
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 2
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
      error |-> TRUE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
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
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> TRUE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |-> [channel |-> "buckybucky", denom |-> "eth", port |-> ""],
      receiver |-> "a2",
      sender |-> "a1"],
  destChannel |-> "",
  destPort |-> "zarkozarko",
  sourceChannel |-> "",
  sourcePort |-> ""]

(* Transition 0 to State6 *)

State6 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
/\ count = 4
/\ error = TRUE
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
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 1
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 2
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
      error |-> TRUE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
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
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> TRUE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 4
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "buckybucky", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "zarkozarko",
          sourceChannel |-> "",
          sourcePort |-> ""]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |->
        [channel |-> "buckybucky", denom |-> "", port |-> "buckybucky"],
      receiver |-> "",
      sender |-> "a1"],
  destChannel |-> "",
  destPort |-> "",
  sourceChannel |-> "buckybucky",
  sourcePort |-> "buckybucky"]

(* Transition 0 to State7 *)

State7 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
/\ count = 5
/\ error = TRUE
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
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 1
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 2
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
      error |-> TRUE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
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
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> TRUE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 4
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "buckybucky", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "",
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
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "buckybucky", denom |-> "", port |-> "buckybucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "buckybucky",
          sourcePort |-> "buckybucky"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "", port |-> ""],
      receiver |-> "a1",
      sender |-> "a2"],
  destChannel |-> "zarkozarko",
  destPort |-> "zarkozarko",
  sourceChannel |-> "",
  sourcePort |-> "zarkozarko"]

(* Transition 0 to State8 *)

State8 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
/\ count = 6
/\ error = TRUE
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
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 1
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarkozarko",
                  denom |-> "atom",
                  port |-> "zarkozarko"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "buckybucky",
          destPort |-> "buckybucky",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 2
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
      error |-> TRUE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
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
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> TRUE,
      handler |-> "OnTimeoutPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarkozarko",
          sourcePort |-> "zarkozarko"]]
  @@ 4
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "buckybucky", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "",
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
            :> 0,
      bankBefore |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0,
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "buckybucky", denom |-> "", port |-> "buckybucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "buckybucky",
          sourcePort |-> "buckybucky"]]
  @@ 6
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
      error |-> TRUE,
      handler |-> "OnRecvPacket",
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "", port |-> ""],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "zarkozarko",
          destPort |-> "zarkozarko",
          sourceChannel |-> "",
          sourcePort |-> "zarkozarko"]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |-> [channel |-> "", denom |-> "", port |-> ""],
      receiver |-> "",
      sender |-> ""],
  destChannel |-> "",
  destPort |-> "",
  sourceChannel |-> "",
  sourcePort |-> "zarkozarko"]

(* The following formula holds true in the last state and violates the invariant *)

InvariantViolation ==
  count = 6
    /\ BMC!Skolem((\E s$3 \in DOMAIN history:
      history[s$3]["handler"] = "OnRecvPacket"))
    /\ BMC!Skolem((\E s$4 \in DOMAIN history:
      history[s$4]["handler"] = "OnTimeoutPacket"))

================================================================================
\* Created by Apalache on Fri Nov 20 12:28:10 CET 2020
\* https://github.com/informalsystems/apalache
