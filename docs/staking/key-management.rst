Key Management
==============

Here we explain a bit how to work with your keys, using the
``gaia client keys`` subcommand. 

**Note:** This keys tooling is not considered production ready and is
for dev only.

We'll look at what you can do using the six sub-commands of
``gaia client keys``:

::

    new
    list
    get
    delete
    recover
    update

Create keys
-----------

``gaia client keys new`` has two inputs (name, password) and two outputs
(address, seed).

First, we name our key:

::

    gaia client keys new alice

This will prompt (10 character minimum) password entry which must be
re-typed. You'll see:

::

    Enter a passphrase:
    Repeat the passphrase:
    alice       A159C96AE911F68913E715ED889D211C02EC7D70
    **Important** write this seed phrase in a safe place.
    It is the only way to recover your account if you ever forget your password.

    pelican amateur empower assist awkward claim brave process cliff save album pigeon intact asset

which shows the address of your key named ``alice``, and its recovery
seed. We'll use these shortly.

Adding the ``--output json`` flag to the above command would give this
output:

::

    Enter a passphrase:
    Repeat the passphrase:
    {
      "key": {
        "name": "alice",
        "address": "A159C96AE911F68913E715ED889D211C02EC7D70",
        "pubkey": {
          "type": "ed25519",
          "data": "4BF22554B0F0BF2181187E5E5456E3BF3D96DB4C416A91F07F03A9C36F712B77"
        }
      },
      "seed": "pelican amateur empower assist awkward claim brave process cliff save album pigeon intact asset"
    }

To avoid the prompt, it's possible to pipe the password into the
command, e.g.:

::

    echo 1234567890 | gaia client keys new fred --output json

After trying each of the three ways to create a key, look at them, use:

::

    gaia client keys list

to list all the keys:

::

    All keys:
    alice       6FEA9C99E2565B44FCC3C539A293A1378CDA7609
    bob     A159C96AE911F68913E715ED889D211C02EC7D70
    charlie     784D623E0C15DE79043C126FA6449B68311339E5

Again, we can use the ``--output json`` flag:

::

    [
      {
        "name": "alice",
        "address": "6FEA9C99E2565B44FCC3C539A293A1378CDA7609",
        "pubkey": {
          "type": "ed25519",
          "data": "878B297F1E863CC30CAD71E04A8B3C23DB71C18F449F39E35B954EDB2276D32D"
        }
      },
      {
        "name": "bob",
        "address": "A159C96AE911F68913E715ED889D211C02EC7D70",
        "pubkey": {
          "type": "ed25519",
          "data": "2127CAAB96C08E3042C5B33C8B5A820079AAE8DD50642DCFCC1E8B74821B2BB9"
        }
      },
      {
        "name": "charlie",
        "address": "784D623E0C15DE79043C126FA6449B68311339E5",
        "pubkey": {
          "type": "ed25519",
          "data": "4BF22554B0F0BF2181187E5E5456E3BF3D96DB4C416A91F07F03A9C36F712B77"
        }
      },
    ]

to get machine readable output.

If we want information about one specific key, then:

::

    gaia client keys get charlie --output json

will, for example, return the info for only the "charlie" key returned
from the previous ``gaia client keys list`` command.

The keys tooling can support different types of keys with a flag:

::

    gaia client keys new bit --type secp256k1

and you'll see the difference in the ``"type": field from``\ gaia client
keys get\`

Before moving on, let's set an enviroment variable to make
``--output json`` the default.

Either run or put in your ``~/.bash_profile`` the following line:

::

    export BC_OUTPUT=json

Recover a key
-------------

Let's say, for whatever reason, you lose a key or forget the password.
On creation, you were given a seed. We'll use it to recover a lost key.

First, let's simulate the loss by deleting a key:

::

    gaia client keys delete alice

which prompts for your current password, now rendered obsolete, and
gives a warning message. The only way you can recover your key now is
using the 12 word seed given on initial creation of the key. Let's try
it:

::

    gaia client keys recover alice-again

which prompts for a new password then the seed:

::

    Enter the new passphrase:
    Enter your recovery seed phrase:
    strike alien praise vendor term left market practice junior better deputy divert front calm
    alice-again CBF5D9CE6DDCC32806162979495D07B851C53451

and voila! You've recovered your key. Note that the seed can be typed
out, pasted in, or piped into the command alongside the password.

To change the password of a key, we can:

::

    gaia client keys update alice-again

and follow the prompts.

That covers most features of the keys sub command.

.. raw:: html

   <!-- use later in a test script, or more advance tutorial?
   SEED=$(echo 1234567890 | gaia client keys new fred -o json | jq .seed | tr -d \")
   echo $SEED
   (echo qwertyuiop; echo $SEED stamp) | gaia client keys recover oops
   (echo qwertyuiop; echo $SEED) | gaia client keys recover derf
   gaia client keys get fred -o json
   gaia client keys get derf -o json
   ```
   -->
