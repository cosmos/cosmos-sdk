# SDK 应用程序剖析

## Node Client

全节点的核心进程是基于 SDK 包的。网络中的参与者运行此过程以初始化其状态机，与其他全节点连接并在新块进入时更新其状态机。

```
                ^  +-------------------------------+  ^
                |  |                               |  |
                |  |  State-machine = Application  |  |
                |  |                               |  |   Built with Cosmos SDK
                |  |            ^      +           |  |
                |  +----------- | ABCI | ----------+  v
                |  |            +      v           |  ^
                |  |                               |  |
Blockchain Node |  |           Consensus           |  |
                |  |                               |  |
                |  +-------------------------------+  |   Tendermint Core
                |  |                               |  |
                |  |           Networking          |  |
                |  |                               |  |
                v  +-------------------------------+  v
```

区块链全节点以二进制形式表示，通常以 `-d` 后缀表示`守护程序`（例如，`appd` 表示 `app` 或 `gaiad` 表示 `gaia`）。这个二进制文件是通过编译一个简单的代码文件 main.go 构建的，`main.go` 通常位于`./cmd/appd/`中。 此操作通常通过用 Makefile 编译。

编译了二进制文件，就可以通过运行[`start`命令](https://docs.cosmos.network/master/core/node.html#start-command) 来启动节点。 此命令功能主要执行三件事：

1. [`app.go`] 创建了一个状态机实例。

2. 用最新的已知状态初始化状态机，该状态机是从存储在 `~/.appd/data` 文件夹中的 db 中提取的。 此时，状态机的高度为：`appBlockHeight`。

3. 创建并启动一个新的 Tendermint 实例。 该节点将与对等节点进行连接交换信息。 它将从他们那里获取最新的 `blockHeight`，如果它大于本地的 `appBlockHeight`，则重播块以同步到该高度。 如果 `appBlockHeight` 为 `0`，则该节点从创世开始，并且 Tendermint 通过 ABCI 接口向 `app` 发送 `InitChain` 初始化链命令，从而触发 [`InitChainer`](https://docs.cosmos.network/master/basics/app-anatomy.html#initchainer)。

## Core Application File

通常，状态机的核心是在名为 `app.go` 的文件中定义的。 它主要包含“应用程序的类型定义”和“创建和初始化它”的功能。

### Type Definition of the Application

在 app.go 中重要的一个是应用程序的 type。 它通常由以下部分组成：

- 在 `app.go` 中定义的自定义应用程序是 `baseapp` 的扩展。 当事务由 Tendermint 发送到应用程序时，`app` 使用 `baseapp` 的方法将它们转送到对应的模块。 baseapp 为应用程序实现了大多数核心逻辑，包括所有的 [ABCI 方法](https://tendermint.com/docs/spec/abci/abci.html#overview)和转送消息逻辑。

- 一条 key 链包含整个状态，他是基于 Cosmos SDK 的 multistore 实现的。 每个模块使用 multistore 的一个或多个存储来存储其状态。可以使用在 `app` 类型中声明的特定键来访问这些存储。这些密钥以及 `keepers` 是 Cosmos SDK 的对象功能模型的核心。
- 模块 `keeper` 的列表。 每个模块都会抽象定义一个 keeper，该 keeper 实现模块存储的读写。 一个模块的 `keeper` 方法可以从其他模块（如果已授权）中调用，这就是为什么它们在应用程序的类型中声明并作为接口导出到其他模块的原因，以便后者只能访问授权的功能。
- 应用程序的 `codec` 用于序列化和反序列化数据结构以便存储它们，因为存储只能持久化 `[]bytes`。 `编解码器`必须是确定性的。 默认编解码器为 amino
- 模块管理器是一个对象，其中包含应用程序模块的列表。 它简化了与这些模块相关的操作，例如注册 routes 操作，query route 操作或设置各种功能的模块之间顺序执行情况，例如 InitChainer 操作，BeginBlocke 操作和 EndBlocker 操作
- 请参阅 [gaia](https://github.com/cosmos/gaia) 中的应用程序类型定义示例

+++ https://github.com/cosmos/gaia/blob/5bc422e6868d04747e50b467e8eeb31ae2fe98a3/app/app.go#L87-L115

### Constructor Function

此函数构造了以上部分中定义的类型的新应用程序。在应用程的 start 命令中使用，它必须具有 AppCreator 签名。

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/server/constructors.go#L20

以下是此功能执行的主要操作：

- 创建初始化一个新的 codec 实例，并使用基础模块管理器初始化每个应用程序模块的 `codec`。
- 使用 baseapp 实例，编解码器和所有适当的存储键的引用实例化一个新应用程序。
- 使用每个应用程序模块的 `NewKeeper` 功能实例化在应用程序的`类型`中定义的所有 keeper。 注意，所有 keeper 必须以正确的顺序实例化，因为一个模块的 NewKeeper 可能需要引用另一个模块的 `keeper`。
- 使用每个应用模块的 AppModule 来实例化应用程序的模块管理器
- 使用模块管理器，初始化应用程序的 routes 和 query route。 当事务由 Tendermint 通过 ABCI 中继到应用程序时，它使用此处定义的路由被路由到相应模块的回调 handler。 同样，当应用程序收到查询时，使用此处定义的查询路由将其路由到适当的模块的 querier。
- 使用模块管理器，注册应用程序的模块的 invariants。 invariants 是在每个块末尾评估的变量（例如 token 的总供应量）。 检查不变式的过程是通过 InvariantsRegistry 的特殊模块完成的。 invariants 应等于模块中定义的预测值。 如果该值与预测的值不同，则将触发不变注册表中定义的特殊逻辑（通常会中断链）。这对于确保不会发现任何严重错误并产生难以修复的长期影响非常有用。
- 使用模块管理器，在每个应用程序的模块 的 InitGenesis，BegingBlocker 和 EndBlocker 函数之间设置执行顺序。 请注意，并非所有模块都实现这些功能。
- 模块实现这些功能。
- 设置其余的应用程序参数：
  - `InitChainer` 于在应用程序首次启动时对其进行初始化。
  - `BeginBlocker`，`EndBlocker`：在每个块的开始和结尾处调用。
  - `anteHandler`：用于处理费用和签名验证。
- 挂载存储.
- 返回应用实例.

请注意，此函数仅创建该应用的一个实例，而如果重新启动节点，则状态将从 `〜/.appd/data` 文件夹中保留下来状态加载，如果节点是第一次启动，则从创世文件生成。See an example of application constructor from [`gaia`](https://github.com/cosmos/gaia):

+++ https://github.com/cosmos/gaia/blob/f41a660cdd5bea173139965ade55bd25d1ee3429/app/app.go#L110-L222

### InitChainer

InitChainer 用于根据创始文件（即创始账户的代币余额）初始化应用程序的状态。 当应用程序从 Tendermint 引擎收到`InitChain`消息时调用该消息，该消息是在节点以`appBlockHeight == 0`（即创世）启动。 应用程序必须通过[`SetInitChainer`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp#BaseApp.SetInitChainer)方法设置其[constructor](https://docs.cosmos.network/master/basics/app-anatomy.html#constructor-function)中的`Initchainer`。

通常，`InitChainer`主要由每个应用程序模块的 InitGenesis 函数组成。 这是通过调用模块管理器的 InitGenesis 函数来完成的，而模块管理器的 InitGenesis 函数将依次调用其包含的每个模块的 InitGenesis 函数。 请注意，必须使用模块管理器的 SetOrderInitGenesis 方法设置模块的 InitGenesis 函数的顺序。 这是在 应用程序的构造函数 application-constructor 中完成的，必须在 SetInitChainer 之前调用 SetOrderInitGenesis。

查看来自[gaia](https://github.com/cosmos/gaia)的 InitChainer 的示例：

See an example of an `InitChainer` from [`gaia`](https://github.com/cosmos/gaia):

查看来自 [`gaia`](https://github.com/cosmos/gaia)的 `InitChainer` 的示例：
+++ https://github.com/cosmos/gaia/blob/f41a660cdd5bea173139965ade55bd25d1ee3429/app/app.go#L235-L239

### BeginBlocker and EndBlocker

该 SDK 为开发人员提供了在其应用程序中实现自定义代码可能性。 这是通过两个名为 `BeginBlocker` 和 `EndBlocker` 的函数实现的。当应用程序分别从 Tendermint 引擎接收到 `BeginBlock` 和 `EndBlock` 消息时，将调用它们，它们分别在每个块的开始和结尾处发生。应用程序必须通过 [SetBeginBlocker](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp) 和 [SetEndBlocker](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp#BaseApp.SetEndBlocker) 方法在其 constructor 中设置 `BeginBlocker` 和 `EndBlocker`。

通常，`BeginBlocker` 和 `EndBlocker` 函数主要由每个应用程序模块的 `BeginBlock` 和 `EndBlock` 函数组成。 这是通过调用模块管理器的 BeginBlock 和 EndBlock 函数来完成的，而后者又会调用其包含的每个模块的 BeginBLock 和 EndBlock 函数。 请注意，必须分别在模块管理器中使用 SetOrderBeginBlock 和 SetOrderEndBlock 方法来设置模块的 BegingBlock 和 EndBlock 函数必须调用的顺序。这是通过应用程序的构造函数中的模块管理器完成的，必须调用 SetOrderBeginBlock 和 SetOrderEndBlock 方法。 在 SetBeginBlocker 和 SetEndBlocker 函数之前。

附带说明，请记住特定于应用程序的区块链是确定性的，这一点很重要。开发人员必须注意不要在 BeginBlocker 或 EndBlocker 中引入不确定性，还必须注意不要使它们在计算上过于昂贵，因为[gas]不会限制计算代价当调用 BeginBlocker 和 EndBlocker 执行。

请参阅 [gaia](https://github.com/cosmos/gaia)中的 `BeginBlocker` 和 `EndBlocker` 函数的示例。

+++ https://github.com/cosmos/gaia/blob/f41a660cdd5bea173139965ade55bd25d1ee3429/app/app.go#L224-L232

### Register Codec

MakeCodec 函数是 app.go 文件的最后一个重要功能。 此函数的目的是使用 RegisterLegacyAminoCodec 函数实例化 codec`cdc`，例如 amino 初始化 SDK 的编解码器以及每个应用程序的模块。

为了注册应用程序的模块，`MakeCodec` 函数在 `ModuleBasics` 上调用 `RegisterLegacyAminoCodec`。`ModuleBasics` 是一个基本管理器，其中列出了应用程序的所有模块。 它在`init()`函数中得到实例化，仅用于注册应用程序模块的非依赖元素（例如编解码器）。 要了解有关基本模块管理器的更多信息，请点击[这里](https://docs.cosmos.network/master/building-modules/module-manager.html#basicmanager)。

请参阅 [gaia](https://github.com/cosmos/gaia) 中的 `MakeCodec` 示例：

+++ https://github.com/cosmos/gaia/blob/f41a660cdd5bea173139965ade55bd25d1ee3429/app/app.go#L64-L70

## Modules

Modules 是 SDK 应用程序的灵魂。它们可以被视为状态机中的状态机。当交易通过 ABCI 从底层的 Tendermint 引擎中继到应用程序时，它由 baseapp 找到对应的模块以便进行处理。这种范例使开发人员可以轻松构建复杂的状态机，因为他们所需的大多数模块通常已经存在。对于开发人员而言，构建 SDK 应用程序所涉及的大部分工作都围绕构建其应用程序尚不存在的自定义模块，并将它们与已经存在的模块集成到一个统一的应用程序中。在应用程序目录中，标准做法是将模块存储在 `x/` 文件夹中（不要与 SDK 的`x/`文件夹混淆，该文件夹包含已构建的模块）。

### Application Module Interface

模块必须实现 Cosmos SDK AppModuleBasic 中的 [interfaces](https://docs.cosmos.network/master/building-modules/module-manager.html#application-module-interfaces) 和 AppModule。 前者实现了模块的基本非依赖性元素，例如`编解码器`，而后者则处理了大部分模块方法（包括需要引用其他模块的`keeper`的方法）。`AppModule` 和 `AppModuleBasic` 类型都在名为 `module.go` 的文件中定义。

AppModule 在模块上公开了一组有用的方法，这些方法有助于将模块组合成一个一致的应用程序。 这些方法是从模块管理器中调用的，该模块管理应用程序的模块集合。

### Message Types

每个 `module` 定义 [messages](https://docs.cosmos.network/master/building-modules/messages-and-queries.html#messages) 接口。 每个 `transaction` 包含一个或多个 `messages`。

当全节点接收到有效的交易块时，Tendermint 通过 [`DeliverTx`](https://tendermint.com/docs/app-dev/abci-spec.html#delivertx) 将每个交易发到应用程序。然后，应用程序处理事务：

- 收到交易后，应用程序首先从 `[]bytes` 反序列化得到。
- 然后，在提取交易中包含的消息之前，它会验证有关交易的一些信息，例如费用支付和签名。
- 使用 message 的 Type()方法，baseapp 可以将其发到对应模块的回调 handler 以便对其进行处理。
- 如果消息已成功处理，则状态将更新。

有关事务生命周期的更多详细信息，请看[这里](./ tx-lifecycle.md)。

模块开发人员在构建自己的模块时会创建自定义消息类型。 通常的做法是在消息的类型声明前加上 `Msg`。 例如，消息类型 `MsgSend` 允许用户传输 tokens：

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/bank/internal/types/msgs.go#L10-L15

它由 `bank` 模块的回调 `handler` 处理，最终会调用 `auth` 模块来写 `keeper` 以更新状态。

### Handler

回调 `handler` 是指模块的一部分，负责处理 `baseapp` 传递的 `message` 消息。 仅当通过 ABCI 接口的 DeliverTx 消息从 Tendermint 收到事务时，才执行模块的`处理程序`功能。 如果通过 CheckTx，仅执行无状态检查和与费用相关的有状态检查。为了更好地理解 `DeliverTx` 和 `CheckTx` 之间的区别以及有状态和无状态检查之间的区别，请看[这里](./tx-lifecycle.md)。

模块的`处理程序`通常在名为 `handler.go` 的文件中定义，并包括：

- NewHandler 将消息发到对应的回调 `handler`。 该函数返回一个 `handler` 函数，此前这个函数在 `AppModule` 中注册，以在应用程序的模块管理器中用于初始化应用程序的路由器。接下来是 [nameservice tutorial](https://github.com/cosmos/sdk-tutorials/tree/master/nameservice) 的一个例子。

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/internal/keeper/querier.go#L19-L32

- 模块定义的每种消息类型的处理函数。开发人员在这些函数中编写消息处理逻辑。这通常包括进行状态检查以确保消息有效，并调用 [`keeper`](https://docs.cosmos.network/master/basics/app-anatomy.html#keeper) 的方法来更新状态。

处理程序函数返回结果类型为 sdk.Result，该结果通知应用程序消息是否已成功处理：

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/result.go#L15-L40

### Querier

`Queriers` 与 `handlers` 非常相似，除了它们向状态查询用户而不是处理事务。 最终用户从 interface 发起 query，最终用户会提供 `queryRoute` 和一些 `data`。 然后使用 `queryRoute` 通过 `baseapp` 的 `handleQueryCustom` 方法查询到正确的应用程序的 `querier` 函数

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/baseapp/abci.go#L395-L453

模块的 Querier 是在名为 querier.go 的文件中定义的，包括：

- NewQuerier 将查找到对应 query 函数。 此函数返回一个 `querier` 函数，此前它在 AppModule 中注册，以在应用程序的模块管理器中用于初始化应用程序的查询路由器。请参阅 [nameservice demo]（https://github.com/cosmos/sdk-tutorials/tree/master/nameservice）中的此类切换示例：
  +++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/internal/keeper/querier.go#L19-L32

- 对于模块定义的每种需要查询的数据类型，都具有一个查询器功能。开发人员在这些函数中编写查询处理逻辑。这通常涉及调用 `keeper` 的方法来查询状态并将其序列化为 JSON。

### Keeper

[`Keepers`](https://docs.cosmos.network/master/building-modules/keeper.html)是其模块存储器件。要在模块的存储区中进行读取或写入，必须使用其 `keeper` 方法之一。这是由 Cosmos SDK 的 object-capabilities 模型确保的。 只有持有存储密钥的对象才能访问它，只有模块的 `keeper` 才应持有该模块存储的密钥。

`Keepers` 通常在名为 `keeper.go` 的文件中定义。 它包含 `keeper` 的类型定义和方法。

`keeper` 类型定义通常包括：

- 多重存储中模块存储的`密钥`。
  - 参考**其他模块的`keepers`**。 仅当 `keeper` 需要访问其他模块的存储（从它们读取或写入）时才需要。
- 对应用程序的`编解码器`的引用。 `keeper` 需要它在存储结构之前序列化处理，或在检索它们时将反序列化处理，因为存储仅接受 `[]bytes` 作为值。

与类型定义一起，keeper.go 文件的一个重要组成部分是 Keeper 的构造函数 NewKeeper。 该函数实例化上面定义的类型的新 `keeper`，并带有 `codec`，存储 `keys` 以及可能引用其他模块的 `keeper` 作为参数。从应用程序的构造函数中调用 `NewKeeper` 函数。文件的其余部分定义了 `keeper` 的方法，主要是 getter 和 setter。

### Command-Line and REST Interfaces

每个模块都定义了 application-interfaces 向用户公开的命令行命令和 REST routes。 用户可以创建模块中定义的类型的消息，或查询模块管理的状态的子集。

#### CLI

通常，与模块有关的命令在模块文件夹中名为 `client/cli` 的文件夹中定义。CLI 将命令分为交易和查询两类，分别在 `client/cli/tx.go` 和 `client/cli/query.go` 中定义。这两个命令基于 [Cobra Library](https://github.com/spf13/cobra)之上：

- Transactions 命令使用户可以生成新的事务，以便可以将它们包含在块中并更新状态。应该为模块中定义的每个消息类型 message-types 创建一个命令。该命令使用户提供的参数调用消息的构造函数，并将其包装到事务中。SDK 处理签名和其他事务元数据的添加。
- 用户可以查询模块定义的状态子集。查询命令将查询转发到应用程序的查询路由器，然后将查询路由到提供的`queryRoute`参数的相应 querier。

#### REST

模块的 REST 接口允许用户生成事务并通过对应用程序的 light client daemon（LCD） 查询状态。 REST 路由在 `client/rest/rest.go` 文件中定义，该文件包括：

- `RegisterRoutes` 函数，用于注册路由。从主应用程序的接口 application-interfaces 中为应用程序内使用的每个模块调用此函数。SDK 中使用的路由器是 [Gorilla's mux](https://github.com/gorilla/mux)。
- 需要公开的每个查询或事务创建功能的自定义请求类型定义。这些自定义请求类型基于 Cosmos SDK 的基本`请求`类型构建：
  +++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/rest/rest.go#L47-L60

- 每个请求的一个处理函数可以找到给定的模块。 这些功能实现了服务请求所需的核心逻辑。

## Application Interface

Interfaces 允许用户与全节点客户端进行交互。 这意味着从全节点查询数据，或者接受全节点中包含在块中的新事务。

通过汇总在应用程序使用的每个模块中定义的 CLI 命令构建 SDK 应用程序的 CLI。 应用程序的 CLI 通常具有后缀-cli（例如 appcli），并在名为`cmd / appcli / main.go`的文件中定义。 该文件包含：

- main()函数，用于构建 appcli 接口客户端。这个函数准备每个命令，并在构建它们之前将它们添加到`rootCmd`中。在 appCli 的根部，该函数添加了通用命令，例如 status，keys 和 config，查询命令，tx 命令和 rest-server。
- 查询命令是通过调用`queryCmd`函数添加的，该函数也在 appcli / main.go 中定义。此函数返回一个 Cobra 命令，其中包含在每个应用程序模块中定义的查询命令（从`main()`函数作为`sdk.ModuleClients`数组传递），以及一些其他较低级别的查询命令，例如阻止或验证器查询。查询命令通过使用 CLI 的命令“ appcli query [query]”来调用。
- 通过调用`txCmd`函数来添加**交易命令**。与`queryCmd`类似，该函数返回一个 Cobra 命令，其中包含在每个应用程序模块中定义的 tx 命令，以及较低级别的 tx 命令，例如事务签名或广播。使用 CLI 的命令`appcli tx [tx]`调用 Tx 命令。
- registerRoutes 函数，在初始化 轻客户端（LCD）时从 main()函数调用。 “ registerRoutes”调用应用程序每个模块的“ RegisterRoutes”功能，从而注册该模块 routes 到 LCD 的查询路由。可以通过运行以下命令“ appcli rest-server”来启动 LCD。

从[nameservice demo](https://github.com/cosmos/sdk-tutorials/tree/master/nameservice)中查看应用程序的主要命令行文件的示例。

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go

## Dependencies and Makefile

因为开发人员可以自由选择其依赖项管理器和项目构建方法。 也就是说，当前最常用的版本控制框架是[`go.mod`](https://github.com/golang/go/wiki/Modules)。 它确保在整个应用程序中使用的每个库都以正确的版本导入。 请参阅[demo](https://github.com/cosmos/sdk-tutorials/tree/master/nameservice)中的示例：

+++ https://github.com/cosmos/sdk-tutorials/blob/c6754a1e313eb1ed973c5c91dcc606f2fd288811/go.mod#L1-L18

为了构建应用程序，通常使用[Makefile](https://en.wikipedia.org/wiki/Makefile)。 Makefile 主要确保在构建应用程序的两个入口点 [`appd`](https://docs.cosmos.network/master/basics/app-anatomy.html#node-client) 和 [`appcli`](https://docs.cosmos.network/master/basics/app-anatomy.html#application-interface) 之前运行 `go.mod`。 请参阅 nameservice demo 中的 Makefile 示例

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/Makefile

## Next

了解有关[交易生命周期](./ tx-lifecycle.md)的更多信息
