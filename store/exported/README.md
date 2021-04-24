## Purpose

This package exists so that we can import "github.com/cosmos/cosmos-sdk/store/internal/proofs" in fuzzers
without complaints about importing internal libraries. It also exists to generate the corpus that's the seed
to the fuzzers.
