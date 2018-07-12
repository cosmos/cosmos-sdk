# Building an App

::: tip
Lotion requires __node v7.6.0__ or higher, and a mac or linux machine.
:::

## Installation
```
$ npm install lotion
```

## Simple App
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

## Learn More

You can learn more about Lotion JS by visiting Lotion on [Github](https://github.com/keppel/lotion).
