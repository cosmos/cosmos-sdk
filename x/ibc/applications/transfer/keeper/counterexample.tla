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
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "zarko", denom |-> "", port |-> "bucky"],
      receiver |-> "",
      sender |-> "a1"],
  destChannel |-> "",
  destPort |-> "",
  sourceChannel |-> "zarko",
  sourcePort |-> ""]

(* Transition 0 to State3 *)

State3 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
/\ count = 1
/\ error = FALSE
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
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
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
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
      receiver |-> "a1",
      sender |-> "a2"],
  destChannel |-> "zarko",
  destPort |-> "bucky",
  sourceChannel |-> "bucky",
  sourcePort |-> "zarko"]

(* Transition 1 to State4 *)

State4 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 1
/\ count = 2
/\ error = FALSE
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
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
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
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "bucky",
          sourcePort |-> "zarko"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
      receiver |-> "a2",
      sender |-> ""],
  destChannel |-> "bucky",
  destPort |-> "zarko",
  sourceChannel |-> "bucky",
  sourcePort |-> "bucky"]

(* Transition 1 to State5 *)

State5 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 1
/\ count = 3
/\ error = FALSE
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
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
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
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "bucky",
          sourcePort |-> "zarko"]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 1,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "bucky",
          sourcePort |-> "bucky"]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |-> [channel |-> "zarko", denom |-> "", port |-> "zarko"],
      receiver |-> "",
      sender |-> ""],
  destChannel |-> "",
  destPort |-> "",
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
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 1
/\ count = 4
/\ error = TRUE
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
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
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
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "bucky",
          sourcePort |-> "zarko"]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 1,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "bucky",
          sourcePort |-> "bucky"]]
  @@ 4
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "zarko"],
              receiver |-> "",
              sender |-> ""],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "",
          sourcePort |-> ""]]
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

InvariantViolation == count = 4

================================================================================
\* Created by Apalache on Mon Nov 16 14:42:50 CET 2020
\* https://github.com/informalsystems/apalache
