// v5 is an empty package that exists because of the group module.
// the group module v2 migration actually migrates the auth module state (replace group policies accounts from module accounts to base accounts).
// the auth state does not migrate if the group module is not enabled.
package v5
