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
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
      receiver |-> "a2",
      sender |-> ""],
  destChannel |-> "zarko",
  destPort |-> "zarko",
  sourceChannel |-> "zarko",
  sourcePort |-> "zarko"]

(* Transition 1 to State3 *)

State3 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 1
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
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            :> 0,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |-> [channel |-> "", denom |-> "", port |-> ""],
      receiver |-> "a1",
      sender |-> "a1"],
  destChannel |-> "",
  destPort |-> "bucky",
  sourceChannel |-> "",
  sourcePort |-> ""]

(* Transition 0 to State4 *)

State4 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 1
/\ count = 2
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
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            :> 0,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |-> [channel |-> "", denom |-> "", port |-> ""],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "bucky",
          sourceChannel |-> "",
          sourcePort |-> ""]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "atom", port |-> "bucky"],
      receiver |-> "a1",
      sender |-> "a2"],
  destChannel |-> "bucky",
  destPort |-> "",
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
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
      denom |-> "atom",
      port |-> ""]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            :> 0,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |-> [channel |-> "", denom |-> "", port |-> ""],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "bucky",
          sourceChannel |-> "",
          sourcePort |-> ""]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "atom",
              port |-> ""]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "", denom |-> "atom", port |-> "bucky"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "",
          sourceChannel |-> "bucky",
          sourcePort |-> "bucky"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "zarko", denom |-> "", port |-> "zarko"],
      receiver |-> "",
      sender |-> "a1"],
  destChannel |-> "zarko",
  destPort |-> "zarko",
  sourceChannel |-> "zarko",
  sourcePort |-> "zarko"]

(* Transition 0 to State6 *)

State6 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
      denom |-> "atom",
      port |-> ""]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            :> 0,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> ""],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |-> [channel |-> "", denom |-> "", port |-> ""],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "bucky",
          sourceChannel |-> "",
          sourcePort |-> ""]]
  @@ 3
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "atom",
              port |-> ""]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "", denom |-> "atom", port |-> "bucky"],
              receiver |-> "a1",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "",
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "atom",
              port |-> ""]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "atom",
              port |-> ""]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "", port |-> "zarko"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
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

InvariantViolation == count = 4 /\ error = TRUE

================================================================================
\* Created by Apalache on Wed Nov 11 14:47:49 CET 2020
\* https://github.com/informalsystems/apalache
