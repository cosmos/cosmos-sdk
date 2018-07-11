/*
Package mock provides functions for creating applications for testing.

This module also features randomized testing, so that various modules can test
that their operations are interoperable.

The intended method of using this randomized testing framework is that every
module provides TestAndRunTx methods for each of its desired methods of fuzzing
its own txs, and it also provides the invariants that it assumes to be true.
You then pick and choose from these tx types and invariants. To pick and choose
these, you first build a mock app with the correct keepers. Then you call the
app.RandomizedTesting method with the set of desired txs, invariants, along
with the setups each module requires.
*/
package mock
