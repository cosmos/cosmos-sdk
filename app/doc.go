/*
Package app contains data structures that provide basic data storage
functionality and act as a bridge between the ABCI interface and the internal
SDK representations.

BaseApp has no state except the MultiStore you provide upon init.  You must
also provide a Handler and a TxParser.
*/
package app
