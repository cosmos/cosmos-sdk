/*
Package simulation implements a simulation framework for any state machine
built on the SDK which utilizes auth.

It is primarily intended for fuzz testing the integration of modules. It will
test that the provided operations are interoperable, and that the desired
invariants hold. It can additionally be used to detect what the performance
benchmarks in the system are, by using benchmarking mode and cpu / mem
profiling. If it detects a failure, it provides the entire log of what was ran.

The simulator takes as input: a random seed, the set of operations to run, the
invariants to test, and additional parameters to configure how long to run, and
misc. parameters that affect simulation speed.

It is intended that every module provides a list of Operations which will
randomly create and run a message / tx in a manner that is interesting to fuzz,
and verify that the state transition was executed as exported. Each module
should additionally provide methods to assert that the desired invariants hold.

Then to perform a randomized simulation, select the set of desired operations,
the weightings for each, the invariants you want to test, and how long to run
it for. Then run simulation.Simulate! The simulator will handle things like
ensuring that validators periodically double signing, or go offline.
*/
package simulation
