Feature: MsgAuthorizeCircuitBreaker
  Circuit breaker actions can be authorized:
  - when the granter is a super-admin
  - when the permissions are valid

  Rule: the granter must be a super-admin

    Example: granter is a super-admin
      Given "acct1" has permission "LEVEL_SUPER_ADMIN"
      When "acct1" attempts to grant "acct2" the permissions
        """
        { "level": "LEVEL_ALL_MSGS" }
        """
      Then expect success

    Example: granter has no permissions
      Given "acct1" has no permissions
      When "acct1" attempts to grant "acct2" the permissions
        """
        { "level": "LEVEL_ALL_MSGS" }
        """
      Then expect an "unauthorized" error

    Example: granter has all msg's permissions
      Given "acct1" has permission "LEVEL_ALL_MSGS"
      When "acct1" attempts to grant "acct2" the permissions
        """
        { "level": "LEVEL_ALL_MSGS" }
        """
      Then expect an "unauthorized" error

  Rule: limit_msg_types must be used with LEVEL_SOME_MSGS

    Example: granting LEVEL_SOME_MSGS with limit_msg_types
      Given "acct1" has permission "LEVEL_ALL_MSGS"
      When "acct1" attempts to grant "acct2" the permissions
        """
        {
         "level": "LEVEL_SOME_MSGS"
         "limit_msg_types": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect success

    Example: granting LEVEL_SOME_MSGS without limit_msg_types
      Given "acct1" has permission "LEVEL_SUPER_ADMIN"
      When "acct1" attempts to grant "acct2" the permissions
        """
        { "level": "LEVEL_SOME_MSGS" }
        """
      Then expect an "invalid request" error

    Example: granting LEVEL_ALL_MSGS with limit_msg_types
      Given "acct1" has permission "LEVEL_SUPER_ADMIN"
      When "acct1" attempts to grant "acct2" the permissions
        """
        {
          "level": "LEVEL_ALL_MSGS",
          "limit_msg_types": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect an "invalid request" error

    Example: attempting to revoke with limit_msg_types
      Given "acct1" has permission "LEVEL_SUPER_ADMIN"
      When "acct1" attempts to revoke "acct2" the permissions
        """
        {
          "level": "LEVEL_NONE_UNSPECIFIED",
          "limit_msg_types": "cosmos.bank.v1beta1.MsgSend"
        }
        """
      Then expect an "invalid request" error

  Rule: permissions can be revoked using LEVEL_NONE_UNSPECIFIED

    Example: revoking permissions
      Given "acct1" has permission "LEVEL_SUPER_ADMIN"
      And "acct2" has permission "LEVEL_ALL_MSGS"
      When "acct1" attempts to revoke "acct2" the permissions
        """
        {
          "level": "LEVEL_NONE_UNSPECIFIED",
        }
        """
      Then expect sucesss
      And expect that "acct2" has no permissions
