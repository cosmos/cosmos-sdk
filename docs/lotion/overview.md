# Overview

Lotion is an alternative to the Cosmos SDK and allows you to create blockchain apps in JavaScript. It aims to make writing new blockchain apps fast and easy by using the ABCI protocol to build on top of Tendermint. Lotion lets you write secure, scalable applications that can easily interoperate with other blockchains on the Cosmos Network using IBC.

Lotion itself is a tiny framework; its true power comes from the network of small, focused modules built upon it. Adding a fully-featured cryptocurrency to your blockchain, for example, takes only a few lines of code.

For more information see the [website](https://lotionjs.com) and [GitHub repo](https://github.com/keppel/lotion), for complete documentation which expands on the following example.

## Building an App

### Installation

::: tip
Lotion requires __node v7.6.0__ or higher, and a mac or linux machine.
:::

```
$ npm install lotion
```

### Simple App

`app.js`:

```js
let lotion = require('lotion')

let app = lotion({
  initialState: {
    count: 0
  }
})

app.use(function (state, tx) {
  if(state.count === tx.nonce) {
    state.count++
  }
})

app.listen(3000)
```

run `node app.js`, then:

```bash
$ curl http://localhost:3000/state
# { "count": 0 }

$ curl http://localhost:3000/txs -d '{ "nonce": 0 }'
# { "ok": true }

$ curl http://localhost:3000/state
# { "count": 1 }
```
