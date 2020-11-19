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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
/\ p = [data |->
    [amount |-> 2,
      denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
      receiver |-> "a2",
      sender |-> "a2"],
  destChannel |-> "bucky",
  destPort |-> "zarko",
  sourceChannel |-> "zarko",
  sourcePort |-> "bucky"]

(* Transition 1 to State3 *)

State3 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 2
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
/\ p = [data |->
    [amount |-> 2,
      denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
      receiver |-> "a2",
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
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 2
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 2
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "bucky",
          sourcePort |-> "zarko"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
      receiver |-> "a1",
      sender |-> "a1"],
  destChannel |-> "zarko",
  destPort |-> "bucky",
  sourceChannel |-> "zarko",
  sourcePort |-> "zarko"]

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
      denom |-> "eth",
      port |-> "bucky"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 2
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 2
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
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
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
      receiver |-> "a1",
      sender |-> ""],
  destChannel |-> "bucky",
  destPort |-> "zarko",
  sourceChannel |-> "zarko",
  sourcePort |-> "zarko"]

(* Transition 1 to State6 *)

State6 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
      denom |-> "eth",
      port |-> "bucky"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 2
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 2
/\ count = 4
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
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
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
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
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
/\ p = [data |->
    [amount |-> 2,
      denomTrace |-> [channel |-> "", denom |-> "eth", port |-> "zarko"],
      receiver |-> "",
      sender |-> "a1"],
  destChannel |-> "",
  destPort |-> "",
  sourceChannel |-> "bucky",
  sourcePort |-> "bucky"]

(* Transition 0 to State7 *)

State7 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
      denom |-> "eth",
      port |-> "bucky"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 2
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 2
/\ count = 5
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
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
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
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
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 5
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> "zarko"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "bucky",
          sourcePort |-> "bucky"]]
/\ p = [data |->
    [amount |-> 1,
      denomTrace |-> [channel |-> "zarko", denom |-> "atom", port |-> "zarko"],
      receiver |-> "a2",
      sender |-> "a1"],
  destChannel |-> "zarko",
  destPort |-> "bucky",
  sourceChannel |-> "zarko",
  sourcePort |-> ""]

(* Transition 0 to State8 *)

State8 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
      denom |-> "eth",
      port |-> "bucky"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 2
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 2
/\ count = 6
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
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
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
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
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 5
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> "zarko"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "bucky",
          sourcePort |-> "bucky"]]
  @@ 6
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "atom", port |-> "zarko"],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |-> [channel |-> "bucky", denom |-> "", port |-> "bucky"],
      receiver |-> "",
      sender |-> "a1"],
  destChannel |-> "",
  destPort |-> "",
  sourceChannel |-> "",
  sourcePort |-> "zarko"]

(* Transition 0 to State9 *)

State9 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
      denom |-> "eth",
      port |-> "bucky"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 2
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 2
/\ count = 7
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
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
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
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
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 5
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> "zarko"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "bucky",
          sourcePort |-> "bucky"]]
  @@ 6
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "atom", port |-> "zarko"],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
  @@ 7
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "bucky", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "",
          sourcePort |-> "zarko"]]
/\ p = [data |->
    [amount |-> 0,
      denomTrace |-> [channel |-> "bucky", denom |-> "", port |-> "bucky"],
      receiver |-> "a1",
      sender |-> ""],
  destChannel |-> "bucky",
  destPort |-> "zarko",
  sourceChannel |-> "",
  sourcePort |-> "zarko"]

(* Transition 0 to State10 *)

State10 ==
/\ bank = <<
    [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
      denom |-> "",
      port |-> ""]
  >>
    :> 0
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
      denom |-> "eth",
      port |-> "bucky"]
  >>
    :> 1
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
      denom |-> "eth",
      port |-> "zarko"]
  >>
    :> 2
  @@ <<
    [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
      denom |-> "atom",
      port |-> "bucky"]
  >>
    :> 2
/\ count = 8
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 1
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
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
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a2",
              sender |-> "a2"],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "bucky"]]
  @@ 2
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "atom", port |-> ""],
              receiver |-> "a2",
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
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
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
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> FALSE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> ""],
              receiver |-> "a1",
              sender |-> ""],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "zarko",
          sourcePort |-> "zarko"]]
  @@ 5
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 2,
              denomTrace |-> [channel |-> "", denom |-> "eth", port |-> "zarko"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "bucky",
          sourcePort |-> "bucky"]]
  @@ 6
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 1,
              denomTrace |->
                [channel |-> "zarko", denom |-> "atom", port |-> "zarko"],
              receiver |-> "a2",
              sender |-> "a1"],
          destChannel |-> "zarko",
          destPort |-> "bucky",
          sourceChannel |-> "zarko",
          sourcePort |-> ""]]
  @@ 7
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "bucky", denom |-> "", port |-> "bucky"],
              receiver |-> "",
              sender |-> "a1"],
          destChannel |-> "",
          destPort |-> "",
          sourceChannel |-> "",
          sourcePort |-> "zarko"]]
  @@ 8
    :> [bankAfter |->
        <<
            [channel |-> "", id |-> "", port |-> ""], [channel |-> "",
              denom |-> "",
              port |-> ""]
          >>
            :> 0
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
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
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a1", port |-> ""], [channel |-> "zarko",
              denom |-> "eth",
              port |-> "bucky"]
          >>
            :> 1
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "bucky",
              denom |-> "eth",
              port |-> "zarko"]
          >>
            :> 2
          @@ <<
            [channel |-> "", id |-> "a2", port |-> ""], [channel |-> "zarko",
              denom |-> "atom",
              port |-> "bucky"]
          >>
            :> 2,
      error |-> TRUE,
      packet |->
        [data |->
            [amount |-> 0,
              denomTrace |->
                [channel |-> "bucky", denom |-> "", port |-> "bucky"],
              receiver |-> "a1",
              sender |-> ""],
          destChannel |-> "bucky",
          destPort |-> "zarko",
          sourceChannel |-> "",
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

InvariantViolation == count = 8

================================================================================
\* Created by Apalache on Thu Nov 19 14:16:35 CET 2020
\* https://github.com/informalsystems/apalache
