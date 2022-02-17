
---- MODULE tlcFolds ---- 

EXTENDS Integers, FiniteSets, Sequences

(*****************************************************************************)
(* The folding operator, used to implement computation over a set.           *)
(* Apalache implements a more efficient encoding than the one below.         *)
(* (from the community modules).                                             *)
(*****************************************************************************)
RECURSIVE FoldSet(_,_,_)
FoldSet( Op(_,_), v, S ) == IF S = {}
                            THEN v
                            ELSE LET w == CHOOSE x \in S: TRUE
                                  IN LET T == S \ {w}
                                      IN FoldSet( Op, Op(v,w), T )

(*****************************************************************************)
(* The folding operator, used to implement computation over a sequence.      *)
(* Apalache implements a more efficient encoding than the one below.         *)
(* (from the community modules).                                             *)
(*****************************************************************************)
RECURSIVE FoldSeq(_,_,_)
FoldSeq( Op(_,_), v, seq ) == IF seq = <<>>
                              THEN v
                              ELSE FoldSeq( Op, Op(v,Head(seq)), Tail(seq) )
====
