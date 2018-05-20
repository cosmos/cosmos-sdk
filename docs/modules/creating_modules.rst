Creating a module
=================

Modules are a key part of the SDK since they add new functionalities to the
``BaseApp`` application template. On a SDK project, modules are hosted within
the ``x`` folder.

Module structure
----------------

In general, all modules follow the same folder structure:

::

  x/
  └── <Module name>
      ├── client/
      │   └── cli/
      │       └── tx.go
      ├── errors.go
      ├── handler.go
      ├── handler_test.go
      ├── keeper.go
      ├── keeper_test.go
      ├── types.go
      ├── types_test.go
      └── wire.go

Errors
^^^^^^

The ``errors.go`` file is meant to create custom ``sdk.Errors`` for each error
that a user may encounter.

::

  package <module_name>

  import (
    sdk "github.com/cosmos/cosmos-sdk/types"
  )

  type CodeType = sdk.CodeType


Then you'll need to define the codes for each error type:

::

  // <module_name> errors reserve 800 ~ 899
  const (
    DefaultCodespace          sdk.CodespaceType = 8
    CodeInvalidDifficulty     CodeType          = 801
    CodeNonexistentDifficulty CodeType          = 802
    CodeNonexistentReward     CodeType          = 803
    CodeNonexistentCount      CodeType          = 804
    CodeInvalidProof          CodeType          = 805
    CodeNotBelowTarget        CodeType          = 806
    CodeInvalidCount          CodeType          = 807
    CodeUnknownRequest        CodeType          = sdk.CodeUnknownRequest
  )

You can set the response message as a parameter or predefine a default message:

::

  func codeToDefaultMsg(code CodeType) string {
  	switch code {
  	case CodeInvalidDifficulty:
  		return "Insuffient difficulty"
  	case CodeNonexistentDifficulty:
  		return "Nonexistent difficulty"
  	case CodeNonexistentReward:
  		return "Nonexistent reward"
  	case CodeNonexistentCount:
  		return "Nonexistent count"
  	case CodeInvalidProof:
  		return "Invalid proof"
  	case CodeNotBelowTarget:
  		return "Not below target"
  	case CodeInvalidCount:
  		return "Invalid count"
  	case CodeUnknownRequest:
  		return "Unknown request"
  	default:
  		return sdk.CodeToDefaultMsg(code)
  	}
  }

::

  func msgOrDefaultMsg(msg string, code CodeType) string {
    if msg != "" {
      return msg
    }
    return codeToDefaultMsg(code)
  }

  func newError(codespace sdk.CodespaceType, code CodeType, msg string) sdk.Error {
    msg = msgOrDefaultMsg(msg, code)
    return sdk.NewError(codespace, code, msg)
  }


Handler
^^^^^^^

Keeper
^^^^^^

Types
^^^^^

Wire
^^^^

The ``wire.go`` file is used to register concrete types on wire codec.
In particular you will need to register every ``Msg`` on the codec.

::

  package <module_name>

  import (
    "github.com/cosmos/cosmos-sdk/wire"
  )

  func RegisterWire(cdc *wire.Codec) {
    cdc.RegisterConcrete(CustomMsg1{}, "cosmos-sdk/CustomMsg1", nil)
    cdc.RegisterConcrete(CustomMsg2{}, "cosmos-sdk/CustomMsg2", nil)
    // ...
    cdc.RegisterConcrete(CustomMsgN{}, "cosmos-sdk/CustomMsgN", nil)
  }

Testing
-------
