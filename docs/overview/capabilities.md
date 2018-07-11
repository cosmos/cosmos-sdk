Overview
========

The SDK design optimizes flexibility and security. The framework is
designed around a modular execution stack which allows applications to
mix and match elements as desired. In addition, all modules are
sandboxed for greater application security.

Framework Overview
------------------

### Object-Capability Model

When thinking about security, it's good to start with a specific threat
model. Our threat model is the following:

    We assume that a thriving ecosystem of Cosmos-SDK modules that are easy to compose into a blockchain application will contain faulty or malicious modules.

The Cosmos-SDK is designed to address this threat by being the
foundation of an object capability system.

    The structural properties of object capability systems favor
    modularity in code design and ensure reliable encapsulation in
    code implementation.

    These structural properties facilitate the analysis of some
    security properties of an object-capability program or operating
    system. Some of these — in particular, information flow properties
    — can be analyzed at the level of object references and
    connectivity, independent of any knowledge or analysis of the code
    that determines the behavior of the objects. As a consequence,
    these security properties can be established and maintained in the
    presence of new objects that contain unknown and possibly
    malicious code.

    These structural properties stem from the two rules governing
    access to existing objects:

    1) An object A can send a message to B only if object A holds a
    reference to B.

    2) An object A can obtain a reference to C only
    if object A receives a message containing a reference to C.  As a
    consequence of these two rules, an object can obtain a reference
    to another object only through a preexisting chain of references.
    In short, "Only connectivity begets connectivity."

See the [wikipedia
article](https://en.wikipedia.org/wiki/Object-capability_model) for more
information.

Strictly speaking, Golang does not implement object capabilities
completely, because of several issues:

-   pervasive ability to import primitive modules (e.g. "unsafe", "os")
-   pervasive ability to override module vars
    <https://github.com/golang/go/issues/23161>
-   data-race vulnerability where 2+ goroutines can create illegal
    interface values

The first is easy to catch by auditing imports and using a proper
dependency version control system like Dep. The second and third are
unfortunate but it can be audited with some cost.

Perhaps [Go2 will implement the object capability
model](https://github.com/golang/go/issues/23157).

#### What does it look like?

Only reveal what is necessary to get the work done.

For example, the following code snippet violates the object capabilities
principle:

    type AppAccount struct {...}
    var account := &AppAccount{
        Address: pub.Address(),
        Coins: sdk.Coins{{"ATM", 100}},
    }
    var sumValue := externalModule.ComputeSumValue(account)

The method "ComputeSumValue" implies a pure function, yet the implied
capability of accepting a pointer value is the capability to modify that
value. The preferred method signature should take a copy instead.

    var sumValue := externalModule.ComputeSumValue(*account)

In the Cosmos SDK, you can see the application of this principle in the
basecoin examples folder.

    // File: cosmos-sdk/examples/basecoin/app/init_handlers.go
    package app

    import (
        "github.com/cosmos/cosmos-sdk/x/bank"
        "github.com/cosmos/cosmos-sdk/x/sketchy"
    )

    func (app *BasecoinApp) initRouterHandlers() {

        // All handlers must be added here.
        // The order matters.
        app.router.AddRoute("bank", bank.NewHandler(app.accountMapper))
        app.router.AddRoute("sketchy", sketchy.NewHandler())
    }

In the Basecoin example, the sketchy handler isn't provided an account
mapper, which does provide the bank handler with the capability (in
conjunction with the context of a transaction run).

### Capabilities systems

TODO:

-   Need for module isolation
-   Capability is implied permission
-   Link to thesis

