# Keys CLI

**WARNING: out-of-date and parts are wrong.... please update**

This is as much an example how to expose cobra/viper, as for a cli itself
(I think this code is overkill for what go-keys needs). But please look at
the commands, and give feedback and changes.

`RootCmd` calls some initialization functions (`cobra.OnInitialize` and `RootCmd.PersistentPreRunE`) which serve to connect environmental variables and cobra flags, as well as load the config file. It also validates the flags registered on root and creates the cryptomanager, which will be used by all subcommands.

## Help info

```
# keys help

Keys allows you to manage your local keystore for tendermint.

These keys may be in any format supported by go-crypto and can be
used by light-clients, full nodes, or any other application that
needs to sign with a private key.

Usage:
  keys [command]

Available Commands:
  get         Get details of one key
  list        List all keys
  new         Create a new public/private key pair
  serve       Run the key manager as an http server
  update      Change the password for a private key

Flags:
      --keydir string   Directory to store private keys (subdir of root) (default "keys")
  -o, --output string   Output format (text|json) (default "text")
  -r, --root string     root directory for config and data (default "/Users/ethan/.tlc")

Use "keys [command] --help" for more information about a command.
```

## Getting the config file

The first step is to load in root, by checking the following in order:

* -r, --root command line flag
* TM_ROOT environmental variable
* default ($HOME/.tlc evaluated at runtime)

Once the `rootDir` is established, the script looks for a config file named `keys.{json,toml,yaml,hcl}` in that directory and parses it.  These values will provide defaults for flags of the same name.

There is an example config file for testing out locally, which writes keys to `./.mykeys`.  You can

## Getting/Setting variables

When we want to get the value of a user-defined variable (eg. `output`), we can call `viper.GetString("output")`, which will do the following checks, until it finds a match:

* Is `--output` command line flag present?
* Is `TM_OUTPUT` environmental variable set?
* Was a config file found and does it have an `output` variable?
* Is there a default set on the command line flag?

If no variable is set and there was no default, we get back "".

This setup allows us to have powerful command line flags, but use env variables or config files (local or 12-factor style) to avoid passing these arguments every time.

## Nesting structures

Sometimes we don't just need key-value pairs, but actually a multi-level config file, like

```
[mail]
from = "no-reply@example.com"
server = "mail.example.com"
port = 567
password = "XXXXXX"
```

This CLI is too simple to warant such a structure, but I think eg. tendermint could benefit from such an approach.  Here are some pointers:

* [Accessing nested keys from config files](https://github.com/spf13/viper#accessing-nested-keys)
* [Overriding nested values with envvars](https://www.netlify.com/blog/2016/09/06/creating-a-microservice-boilerplate-in-go/#nested-config-values) - the mentioned outstanding PR is already merged into master!
* Overriding nested values with cli flags? (use `--log_config.level=info` ??)

I'd love to see an example of this fully worked out in a more complex CLI.

## Have your cake and eat it too

It's easy to render data different ways.  Some better for viewing, some better for importing to other programs.  You can just add some global (persistent) flags to control the output formatting, and everyone gets what they want.

```
# keys list -e hex
All keys:
betty   d0789984492b1674e276b590d56b7ae077f81adc
john    b77f4720b220d1411a649b6c7f1151eb6b1c226a

# keys list -e btc
All keys:
betty   3uTF4r29CbtnzsNHZoPSYsE4BDwH
john    3ZGp2Md35iw4XVtRvZDUaAEkCUZP

# keys list -e b64 -o json
[
  {
    "name": "betty",
    "address": "0HiZhEkrFnTidrWQ1Wt64Hf4Gtw=",
    "pubkey": {
      "type": "secp256k1",
      "data": "F83WvhT0KwttSoqQqd_0_r2ztUUaQix5EXdO8AZyREoV31Og780NW59HsqTAb2O4hZ-w-j0Z-4b2IjfdqqfhVQ=="
    }
  },
  {
    "name": "john",
    "address": "t39HILIg0UEaZJtsfxFR62scImo=",
    "pubkey": {
      "type": "ed25519",
      "data": "t1LFmbg_8UTwj-n1wkqmnTp6NfaOivokEhlYySlGYCY="
    }
  }
]
```
