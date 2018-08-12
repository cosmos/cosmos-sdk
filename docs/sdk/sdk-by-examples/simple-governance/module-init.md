## Module initialization 

First, let us go into the module's folder and create a folder for our module.

```bash
cd x/
mkdir simple_governance
cd simple_governance
mkdir -p client/cli client/rest
touch client/cli/simple_governance.go client/rest/simple_governance.go errors.go handler.go handler_test.go keeper_keys.go keeper_test.go keeper.go test_common.go test_types.go types.go wire.go
```

Let us start by adding the files we will need. Your module's folder should look something like that:

```
x
└─── simple_governance
      ├─── client
      │     ├───  cli
      │     │     └─── simple_governance.go
      │     └─── rest
      │           └─── simple_governance.go
      ├─── errors.go
      ├─── handler.go
      ├─── keeper_keys.go
      ├─── keeper.go
      ├─── types.go
      └─── wire.go
```

Let us go into the detail of each of these files.