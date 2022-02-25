// Package ormmocks contains generated mocks for orm types that can be used
// in testing. Right now, this package only contains a mock for ormtable.ValidateHooks
// as this useful way for unit testing using an in-memory database. Rather
// than attempting to mock a whole table or database instance, instead
// a mock Hook instance can be passed in to verify that the expected
// insert/update/delete operations are happening in the database.
//
// The Eq function gomock.Matcher that compares protobuf messages can
// be used in gomock EXPECT functions.
//
// See TestHooks in ormdb/module_test.go for examples of how to use
// mock ValidateHooks in a real-world scenario.
package ormmocks
