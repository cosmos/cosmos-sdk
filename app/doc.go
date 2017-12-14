/*
Package app contains data structures that provide basic data storage
functionality and act as a bridge between the ABCI interface and the SDK
abstractions.

BaseApp has no state except the CommitMultiStore you provide upon init.  You must
also provide a Handler.

Transaction parsing is typically handled by the first Decorator.
*/
package app
