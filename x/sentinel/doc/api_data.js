define({ "api": [
  {
    "type": "post",
    "url": "/send-sign",
    "title": "To Create sigature of the client.",
    "name": "CreateSignature",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>AccountName of the client.</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "password",
            "description": "<p>password of account.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "session_id",
            "description": "<p>session-id.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "amount",
            "description": "<p>Amount to create signature.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "counter",
            "description": "<p>Counter value of the sigature.</p>"
          },
          {
            "group": "Parameter",
            "type": "Boolean",
            "optional": false,
            "field": "isfial",
            "description": "<p>boolean value for is this final signature or not.</p>"
          }
        ]
      }
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "10lz2f928xpzsyggqhc9mu80qj59vx0rc6sedxmsfhca8ysuhhtgqypar3h4ty0pgftwqygp6vm54drttw5grlz4p5n238cvzxe2vpxmu6hhnqvt0uxstg7et4vdqhm4v",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "post",
    "url": "/vpn/getpayment",
    "title": "To get payment of vpn service",
    "name": "GetVPNPayment",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "amount",
            "description": "<p>Amount to send VPN node.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "session_id",
            "description": "<p>session-id.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "counter",
            "description": "<p>Counter value.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Account name of client.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "gas",
            "description": "<p>gas value.</p>"
          },
          {
            "group": "Parameter",
            "type": "Boolean",
            "optional": false,
            "field": "isfinal",
            "description": "<p>is this final signature or not.</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "password",
            "description": "<p>password of account.</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "sign",
            "description": "<p>signature of the client.</p>"
          }
        ]
      }
    },
    "error": {
      "fields": {
        "Error 4xx": [
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "InvalidSessionId",
            "description": "<p>SessionId is invalid</p>"
          },
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "SignatureVerificationFailed",
            "description": "<p>Invalid signature</p>"
          }
        ]
      },
      "examples": [
        {
          "title": "InvalidSessionId-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      6\nABCICode:  65545\nError:     --= Error =--\nData: common.FmtError{format:\"Invalid session Id\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        },
        {
          "title": "SignatureVerificationFailed-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      6\nABCICode:  65545\nError:     --= Error =--\nData: common.FmtError{format:\"signature verification failed\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n   \"Success\": true,\n   \"Hash\": \"629F4603A5A4DE598B58DC494CCC38DB9FD96604\",\n   \"Height\": 353,\n   \"Data\":\"eyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YW9==\",\n   \"Tags\": [\n       {\n           \"key\": \"VnBuIFByb3ZpZGVyIEFkZHJlc3M6\",\n           \"value\": \"Y29zbW9zYWNjYWRkcjF1ZG50Z3pzemVzbjd6M3htNjRoYWZ2amxlZ3JoMzh1a3p3OW03Zw==\"\n       },\n       {\n           \"key\": \"c2Vlc2lvbklk\",\n           \"value\": \"WVZJRW81Y0dIczdkb09UVzRDTk4=\"\n       }\n   ]\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "post",
    "url": "/refund",
    "title": "To Refund the balance of client.",
    "name": "Refund",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>AccountName of the client.</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "password",
            "description": "<p>password of account.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "session_id",
            "description": "<p>session-id.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "gas",
            "description": "<p>Gas value.</p>"
          }
        ]
      }
    },
    "error": {
      "fields": {
        "Error 4xx": [
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "TimeInvalidError",
            "description": "<p>Time is not more than 24 hours</p>"
          },
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "InvalidSessionIdError",
            "description": "<p>SessionId is invalid</p>"
          }
        ]
      },
      "examples": [
        {
          "title": "TimeInvalidError-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      2\nABCICode:  6551245\nError:     --= Error =--\nData: common.FmtError{format:\"time is less than 24 hours  or the balance is negative or equal to zero\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        },
        {
          "title": "InvalidSessionIdError-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      6\nABCICode:  124545\nError:     --= Error =--\nData: common.FmtError{format:\"Invalid SessionId\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n\t{\n  \"Success\": true,\n  \"Hash\": \"868B602828FA48F1D4A03D9D066EB42DEC483AA0\",\n  \"Height\": 1092,\n  \"Data\": \"Qwi/dQ1h0GcdrppVOeyJ0eXBlIjoic2VudGluZWwvcmVnaXN0yJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn192hhGfJVl3g=\",\n  \"Tags\": [\n{\n          \"key\": \"Y2xpZW50IFJlZnVuZCBBZGRyZXNzOg==\",\n          \"value\": \"Y29zbW9zYWNjYWRkcjFndnl0N2FnZHY4Z3h3OGR3bmYybms2cnByOGU5dDltY3hkeGV3cA==\"\n      }\n  ]\n}\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "delete",
    "url": "/master",
    "title": "To Delete Master Node.",
    "name": "deleteMasterNode",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "address",
            "description": "<p>Address of Master Node which we want to delete.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>AccountName of the person who is deleting the Master node.</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "password",
            "description": "<p>password of account.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "gas",
            "description": "<p>Gas value.</p>"
          }
        ]
      }
    },
    "error": {
      "fields": {
        "Error 4xx": [
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "AccountNotExists",
            "description": "<p>Master Node not exists</p>"
          }
        ]
      },
      "examples": [
        {
          "title": "AccountNotExists-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      13\nABCICode:  1245197s\nError:     --= Error =--\nData: common.FmtError{format:\"Account is not exist\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n  \"Success\": true,\n  \"Hash\": \"32EF9DFB6BC24D3159A8310F1AE438BED479466E\",\n  \"Height\": 3698,\n  \"Data\": \"FRTjZrQKAswn4UeyJ0eXBlIwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19Tb1W/Usl/KB3iflg==\",\n  \"Tags\": [\n      {\n          \"key\": \"ZGVsZXRlZCBWcG4gYWRkcmVzcw==\",\n          \"value\": \"42a0CgLMJ+FE29Vv1LJfygd4n5Y=\"\n     }\n ]\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "delete",
    "url": "/vpn",
    "title": "To Delete VPN Node.",
    "name": "deleteVpnNode",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "address",
            "description": "<p>Address of VPN Node which we want to delete.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>AccountName of the person who is deleting the VPN node.</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "password",
            "description": "<p>password of account.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "gas",
            "description": "<p>Gas value.</p>"
          }
        ]
      }
    },
    "error": {
      "fields": {
        "Error 4xx": [
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "AccountNotExists",
            "description": "<p>VPN Node not exists</p>"
          }
        ]
      },
      "examples": [
        {
          "title": "AccountNotExists-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      13\nABCICode:  1245197\nError:     --= Error =--\nData: common.FmtError{format:\"Account is not exist\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n  \"Success\": true,\n  \"Hash\": \"32EF9DFB6BC24D3159A8310F1AE438BED479466E\",\n  \"Height\": 3698,\n  \"Data\": \"FRTjZrQKAswn4UTeyJ0eXBlIjoic2VudGluZWWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19b1W/Usl/KB3iflg==\",\n  \"Tags\": [\n      {\n          \"key\": \"ZGVsZXRlZCBWcG4gYWRkcmVzcw==\",\n          \"value\": \"42a0CgLMJ+FE29Vv1LJfygd4n5Y=\"\n     }\n ]\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "post",
    "url": "/keys",
    "title": "To get account.",
    "name": "getAccount",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Name Account holder name.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "password",
            "description": "<p>Password password for account.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "seed",
            "description": "<p>Seed seed words to get account.</p>"
          }
        ]
      }
    },
    "error": {
      "fields": {
        "Error 4xx": [
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "AccountAlreadyExists",
            "description": "<p>AccountName is  already exists</p>"
          },
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "AccountSeedsNotEnough",
            "description": "<p>Seed words are not enough</p>"
          }
        ]
      },
      "examples": [
        {
          "title": "AccountAlreadyExists-Response:",
          "content": "{\n  Account with name XXXXX... already exists.\n}",
          "type": "json"
        },
        {
          "title": "AccountSeedsNotEnough-Response:",
          "content": "{\n recovering only works with XXX word (fundraiser) or 24 word mnemonics, got: XX words\n}",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n   \"name\": \"vpn\",\n   \"type\": \"local\",\n   \"address\": \"cosmosaccaddr1udntgzszesn7z3xm64hafvjlegrh38ukzw9m7g\",\n   \"pub_key\": \"cosmosaccpub1addwnpepqfjqadxwa9p8tvwhydsakyvkajxgyd0ycanv25u7yff7lqtkwuk8vqcy5cg\",\n   \"seed\": \"hour cram bike donor script fragile together derive capital joy glance morning athlete special hint scrub guitar view popular dream idle inquiry transfer often\"\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "get",
    "url": "/keys/seed",
    "title": "To get seeds for generate keys.",
    "name": "getSeeds",
    "group": "Sentinel_Tendermint",
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\ngarden sunset night final child popular fall ostrich amused diamond lift stool useful brisk very half rice evil any behave merge shift ring chronic\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "get",
    "url": "/session/{sessionId}",
    "title": "To get session data.",
    "name": "getSessionData",
    "group": "Sentinel_Tendermint",
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n   \"name\": \"vpn\",{\n   \"TotalLockedCoins\": [\n       {\n           \"denom\": \"sentinel\",\n           \"amount\": \"10000000000\"\n       }\n   ],\n   \"ReleasedCoins\": [\n       {\n           \"denom\": \"sentinel\",\n           \"amount\": \"5000000000\"\n       }\n   ],\n   \"Counter\": 1,\n   \"Timestamp\": 1537361017,\n   \"VpnPubKey\": [2,97,15,10,206,154,217,19,35,137,55,116,142,249,18,94,82,184,186,222,255,183,15,37,229,108,32,62,209,252,247,182,145],\n   \"CPubKey\": [3,157,182,213,107,56,95,22,24,197,116,75,236,23,60,131,180,160,198,244,216,103,74,189,19,147,141,25,242,109,176,252,39],\n   \"CAddress\": \"cosmosaccaddr130q3n8kkpa9flav0sa5lefjunmruhchg5z6pzd\",\n\t    \"Status\": 1\n\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/query.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "post",
    "url": "/vpn/pay",
    "title": "To Pay for VPN service.",
    "name": "payVPN_service",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "amount",
            "description": "<p>Amount to pay for vpn service.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "vaddress",
            "description": "<p>Address of the vpn service provider.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Account name of Client</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "password",
            "description": "<p>password of account.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "gas",
            "description": "<p>Gas value.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "sig_name",
            "description": "<p>NewAccountName.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "sig_password",
            "description": "<p>NewAccountPassword.</p>"
          }
        ]
      }
    },
    "error": {
      "fields": {
        "Error 4xx": [
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "AccountNotExists",
            "description": "<p>VPN Node not exists</p>"
          },
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "AccountNameAlreadyExists",
            "description": "<p>The new account name is already exist</p>"
          },
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "InsufficientFunds",
            "description": "<p>Funds are less than 100</p>"
          }
        ]
      },
      "examples": [
        {
          "title": "AccountVPNNotExists-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 1\nCode:      9\nABCICode:  65545\nError:     --= Error =--\nData: common.FmtError{format:\"VPN address is not registered\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        },
        {
          "title": "AccountNameAlreadyExists-Response:",
          "content": "{\n\" Account with name XXXXXX already exists.\"\n}",
          "type": "json"
        },
        {
          "title": "InsufficientFunds-Response:",
          "content": "{\n\"Funds must be Greaterthan or equals to 100\"\n}",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n  \"Success\": true,\n  \"Hash\": \"D2C58CAFC580CC39A4CFAB4325991A9378AFE77D\",\n  \"Height\": 1196,\n  \"Data\": \"IjNwWGdHazB5MnBGceyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19TdZdWIwak5xIg==\",\n  \"Tags\": [\n     {\n      \"key\": \"c2VuZGVyIGFkZHJlc3M=\",\n      \"value\": \"Y29zbW9zYWNjYWRkcjFuY3hlbGpjcjRnOWhzdmw3amRuempkazNyNzYyamUzenk4bXU5MA==\"\n     },\n    {\n     \"key\": \"c2Vlc2lvbiBpZA==\",\n     \"value\": \"M3BYZ0drMHkycEZxN1l1YjBqTnE=\"\n    }\n         ]\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "post",
    "url": "/register/master",
    "title": "To register Master Node.",
    "name": "registerMasterNode",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Account name of Master Node.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "gas",
            "description": "<p>Gas value.</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "password",
            "description": "<p>password of account.</p>"
          }
        ]
      }
    },
    "error": {
      "fields": {
        "Error 4xx": [
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "AccountAlreadyExists",
            "description": "<p>Master Node already exists</p>"
          }
        ]
      },
      "examples": [
        {
          "title": "AccountAlreadyExists-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      13\nABCICode:  1245197\nError:     --= Error =--\nData: common.FmtError{format:\"Address already Registered as VPN node\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n{\n  \"Success\": true,\n   \"Hash\": \"CF8E073D624F7FA6A41C3CAD9B4A1DB693234225\",\n   \"Height\": 343,\n   \"Data\": \"eyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19==\",\n   \"Tags\": [\n       {\n            \"key\": \"dnBuIHJlZ2lzdGVyZWQgYWRkcmVzcw==\",\n            \"value\": \"Y29zbW9zYWNjYWRkcjFlZ3RydjdxdGU0NnY2cXEzN3p0YzB2dzRuMmhrejZuempycDVhZQ==\"\n        }\n            ]\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "post",
    "url": "/register/vpn",
    "title": "To register VPN service provider.",
    "name": "registerVPN",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "ip",
            "description": "<p>Ip address of VPN service provider.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "upload_speed",
            "description": "<p>Upload Net speed of VPN service.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "download_speed",
            "description": "<p>Download Net speed of VPN service.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "price_per_gb",
            "description": "<p>Price per GB.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "enc_method",
            "description": "<p>Encryption method.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "location_latitude",
            "description": "<p>Latitude Location of service provider.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "location_longitude",
            "description": "<p>Longiude Location of service provider.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "location_city",
            "description": "<p>City Location of service provider.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "location_country",
            "description": "<p>Country Location of service provider.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "node_type",
            "description": "<p>Node type.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "version",
            "description": "<p>version.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Account name of service provider.</p>"
          },
          {
            "group": "Parameter",
            "type": "string",
            "optional": false,
            "field": "password",
            "description": "<p>password of account.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "gas",
            "description": "<p>Gas value.</p>"
          }
        ]
      }
    },
    "error": {
      "fields": {
        "Error 4xx": [
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "AccountAlreadyExists",
            "description": "<p>VPN service provider already exists</p>"
          },
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "NetSpeedInvalidError",
            "description": "<p>Netspeed is Invalid</p>"
          },
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "IpAddressInvalidError",
            "description": "<p>IP address is Invalid</p>"
          },
          {
            "group": "Error 4xx",
            "optional": false,
            "field": "Price_per_GBInvalidError",
            "description": "<p>Price per GB is Invalid</p>"
          }
        ]
      },
      "examples": [
        {
          "title": "AccountAlreadyExists-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      13\nABCICode:  1245197\nError:     --= Error =--\nData: common.FmtError{format:\"Address already Registered as VPN node\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        },
        {
          "title": "NetSpeedInvalidError-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      13\nABCICode:  1245197\nError:     --= Error =--\nData: common.FmtError{format:\"NetSpeed is not Valid\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        },
        {
          "title": "IpAddressInvalidError-Response:",
          "content": "{\n\"  invalid Ip address.\"\n}",
          "type": "json"
        },
        {
          "title": "Price_per_GBInvalidError-Response:",
          "content": "{\ncheckTx failed: (1245197) Msg 0 failed: === ABCI Log ===\nCodespace: 19\nCode:      13\nABCICode:  1245197\nError:     --= Error =--\nData: common.FmtError{format:\"Price per GB is not Valid\", args:[]interface {}(nil)}\nMsg Traces:\n--= /Error =--\n\n=== /ABCI Log ===\n}",
          "type": "json"
        }
      ]
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n  \"Success\": true,\n  \"Hash\": \"CF8E073D624F7FA6A41C3CAD9B4A1DB693234225\",\n  \"Height\": 343,\n  \"Data\": \"eyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19\",\n   \"Tags\": [\n       {\n           \"key\": \"dnBuIHJlZ2lzdGVyZWQgYWRkcmVzcw==\",\n           \"value\": \"Y29zbW9zYWNjYWRkcjFlZ3RydjdxdGU0NnY2cXEzN3p0YzB2dzRuMmhrejZuempycDVhZQ==\"\n       }\n\t\t    ]\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  },
  {
    "type": "post",
    "url": "/send",
    "title": "To send money to account.",
    "name": "sendTokens",
    "group": "Sentinel_Tendermint",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Name Account holder name.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "password",
            "description": "<p>Password password for account.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "to",
            "description": "<p>To address.</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "amount",
            "description": "<p>Amount to send.</p>"
          },
          {
            "group": "Parameter",
            "type": "Number",
            "optional": false,
            "field": "gas",
            "description": "<p>gas value.</p>"
          }
        ]
      }
    },
    "success": {
      "examples": [
        {
          "title": "Response:",
          "content": "{\n  \"Success\": true,\n  \"Hash\": \"CF8E073D624F7FA6A41C3CAD9B4A1DB693234225\",\n  \"Height\": 343,\n  \"Data\": \"eyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19\",\n   \"Tags\": [\n       {\n           \"key\": \"dnBuIHJlZ2lzdGVyZWQgYWRkcmVzcw==\",\n           \"value\": \"Y29zbW9zYWNjYWRkcjFlZ3RydjdxdGU0NnY2cXEzN3p0YzB2dzRuMmhrejZuempycDVhZQ==\"\n       }\n\t\t    ]\n}",
          "type": "json"
        }
      ]
    },
    "version": "0.0.0",
    "filename": "rest/service.go",
    "groupTitle": "Sentinel_Tendermint"
  }
] });
