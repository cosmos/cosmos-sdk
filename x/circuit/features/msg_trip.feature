Feature: MsgTripCircuitBreaker

	Circuit breaker can disable message execution:
	- when the caller trips the circuitbreaker for a message or messages
	- when the permissions are valid

	Rule: a user with LEVEL_SUPER_ADMIN can trip
		Example: caller is a super-admin
			Given "acct1" has permission "LEVEL_SUPER_ADMIN"
			When  attempts to disable msg execution
			"""
			{
				"msg": "cosmos.bank.v1beta1.MsgSend"
			}
			"""
			Then expect success
		Example: caller has no permissions
			Given "acct1" has no permissions
			When "acct1" attempts to disable msg execution
			"""
			{
				"msg": "cosmos.bank.v1beta1.MsgSend"
			}
			"""
			Then expect an "unauthorized" error

	Rule: a user with LEVEL_ALL_MSGS can trip
		Example: caller has all msg permissions
			Given "acct1" has permission "LEVEL_ALL_MSGS"
			When "acct1" attempts to disable msg execution
			"""
			{
				"msg": "cosmos.bank.v1beta1.MsgSend"
			}
			"""
			Then expect success
		Example: caller has no permissions
			Given "acct1" has no permissions
			When "acct1" attempts to disable msg execution
			"""
			{
				"msg": "cosmos.bank.v1beta1.MsgSend"
			}
			"""
			Then expect an "unauthorized" error

	Rule: a user with LEVEL_SOME_MSGS can trip
		Example: caller does not have permission for the list of messages
			Given "acct1" has permission to disable "cosmos.bank.v1beta1.MsgSend" and "cosmos.staking.v1beta1.MsgDelegate"
			When "acct1" attempts to disable msg execution
			"""
			{
			"msgs": ["cosmos.bank.v1beta1.MsgSend",cosmos.staking.v1beta1.MsgDelegate"]
			}
			"""
			Then expect success
		Example: caller does not have permission for the a message in the list of messages
			Given "acct1" has permission to diable "cosmos.bank.v1beta1.MsgSend"
			When "acct1" attempts to disable msg execution
			"""
			{
			"msgs": ["cosmos.bank.v1beta1.MsgSend",cosmos.staking.v1beta1.MsgCreateValidator"]
			}
			"""
			Then expect an "unauthorized" error
		Example: caller has LEVEL_SOME_MSGS permissions
			Given "acct1" has permission to diable "cosmos.bank.v1beta1.MsgSend"
			When "acct1" attempts to disable msg execution
			"""
			{
				"msg": "cosmos.bank.v1beta1.MultiSend"
			}
			"""
			Then expect an "unauthorized" error
