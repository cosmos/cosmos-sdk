## Cosmos SDK design overview 

Cosmos SDK 는 텐더민트 상에서 돌아가는 안전한 상태 머신의 개발을 촉진시켜주는 프레임워크입니다. 이 코어에서, SDK 는 Go 언어로 작성된 ABCI 의 boilerplate 구현체입니다. `multiscore` 는 데이터를 지속시키는데 사용되고 `router` 는 거래를 처리하는데 사용됩니다. 

Cosmos SDK 로 구현된 어플리케이션이 `DeliverTx` 를 통해서 텐더민트으로부터 넘어온 거래를 어떻게 처리하는지 간단하게 설명해보겠습니다 : 

1. 텐더민트 합의 엔진으로 부터 받은 `transaction` 을 deocde 합니다. (기억하십시오. 텐더민트 `[]bytes`의 형태선에서만 처리합니다. 그 내용은 무엇인지 인지하지 않습니다.)
2. `transaction`로 부터 `messages` 를 추출하여 기본적인 온전성을 검사합니다. 
3. 잘 처리될 수 있도록 각 메세지를 적절한 module 에 route 해줍니다. 
4. 상태 변환을 커밋합니다. 

어플리케이션이 당신으로 하여금 거래를 생성하게 할 수도 있습니다, 그들을 encode 하고 해당 텐더민트 엔진에 넘겨 그들을 브로드캐스트합니다. 

### `baseapp`

`baseapp` 은 Cosmos SDK, ABCI 의 boilerplate 구현체입니다. `router` 는 거래를 그들과 관련있는 모듈에게 넘겨주는 역할을 합니다. 당신의 어플리케이션의 메인 파알 `app.go` 은 당신의 `app` 타입을 정의할 것이고 `baseapp` 을 임베트(embed) 할 것입니다. 이 방법으로, 당신의 `app` 타입은 자동적으로 `baseapp` 의 ABCI 메소드 전부를 상속할 것입니다. 그 예시는 [SDK 어플리케이션 튜토리얼](https://github.com/cosmos/sdk-application-tutorial/blob/master/app.go#L27)에서 볼 수 있습니다. 

`baseapp` 의 목표는 저장소와 확장 상태 머신간에 안전한 인터페이스를 제공하고 가능한 상태 머신에 대해서는 최대한 적게 정의합니다.(ABCI의 기능에 충실합니다.)

`baseapp` 에 대한 추가적인 정보는 [여기](https://cosmos.network/docs/concepts/baseapp.html)에서 볼 수 있습니다. 

### Multistore 

Cosmos SDK 는 multiscore 에 지속적인 상태를 제공합니다. multiscore 는 개발자로 하여금 다수의 `KVStores`를 선언할 수 있게 해줍니다. 이 `KVStore` 들은 오직 `[]byte` 타입의 값들만 승인합니다. 그러므로 모든 종류의 구조체는 저장되기 전에 [go-amino](https://github.com/tendermint/go-amino) 로 정렬화 과정을 거쳐야 합니다. 

Multistore 추상화는 각자의 모듈에 의해서 관리되는 고유의 구획에 있는 상태들을 나눌 때 사용다. multistore 에 대한 추가적인 정보를 보려면 [여기](https://cosmos.network/docs/concepts/store.html)를 클릭하십시오. 

### Modules 

Cosmos SDK 의 힘은 그 모듈화에서 옵니다. SDK 어플리케이션은 상호작용하는 모듈들의 모음을 합쳐서 만들어집니다. 각 모듈은 상태의 subset 을 정의하고 그들 고유의 message/transaction 과정을 가지고 있습니다, SDK 는 각 메세지를 이에 어울리는 모듈에 라우팅하는데 사용됩니다. 

아래에 어플리케이션의 각 풀노드가 유효한 블록을 받았을 때 이를 어떻게 처리하는지를 설명하는 과정이 묘사되어있습니다. 

```
+
                                      |
                                      |  Transaction relayed from the full-node's Tendermint engine 
                                      |  to the node's application via DeliverTx
                                      |  
                                      |
                                      |
                +---------------------v--------------------------+
                |                 APPLICATION                    |
                |                                                |
                |     Using baseapp's methods: Decode the Tx,    |
                |     extract and route the message(s)           |
                |                                                |
                +---------------------+--------------------------+
                                      |
                                      |
                                      |
                                      +---------------------------+
                                                                  |
                                                                  |
                                                                  |
                                                                  |  Message routed to the correct
                                                                  |  module to be processed
                                                                  |
                                                                  |
+----------------+  +---------------+  +----------------+  +------v----------+
|                |  |               |  |                |  |                 |
|  AUTH MODULE   |  |  BANK MODULE  |  | STAKING MODULE |  |   GOV MODULE    |
|                |  |               |  |                |  |                 |
|                |  |               |  |                |  | Handles message,|
|                |  |               |  |                |  | Updates state   |
|                |  |               |  |                |  |                 |
+----------------+  +---------------+  +----------------+  +------+----------+
                                                                  |
                                                                  |
                                                                  |
                                                                  |
                                       +--------------------------+
                                       |
                                       | Return result to Tendermint
                                       | (0=Ok, 1=Err)
                                       v
```

각 모듈은 작은 상태머신으로 보일 수도 있습니다. 커스텀화된 메세지 타입들이 상태를 변화시키는만큼 개발자들은 모듈에 의해서 다뤄지는 상태의 subset 을 정의할 필요가 있습니다. (알림: `messages` 는 baseapp을 이용하여 `transactions` 으로 부터 추출된다.) 보통, 각 모듈은 그들만의 `KVStore` multiscore 에 선언하고 그것이 정의하는 상태의 subset 을 유지합니다. 대부분의 개발자들은 자신만의 모듈을 개발할 때 제 3의 모듈에 접근해야 합니다. Cosmos-SDK 는 개방형 프레임워크이기 때문에, 모듈 중에서 일부는 악의적인 의도를 갖고 있을 수도 있습니다. 즉, 모듈간 상호운용을 하는 과정에서 측정할 수 있는 보안 원칙들이 필요하다는 것입니다. 이 원칙들은 [ocap](https://cosmos.network/docs/intro/ocap.html) 을 기반으로 하고 있습니다. 이는 즉, 각 모듈이 다른 모듈들에 접근권한을 가지는 것이 아닌, 각 모듈이 keeper 라는 특별한 object 를 만들어 다른 모듈들에 넘겨 사전에 정의된 기능들을 실행시큰 형식으로 진행된다는 것을 의미합니다.

SDK 모듈들은 SDK의 `x/` 폴더 안에 정의되어 있습니다. 몇몇 핵심 모듈은 아래의 것들을 포함하고 있습니다 : 

- `x/auth` : 계정과 서명을 관리하는데 사용됩니다. 
- `x/bank` : 토큰을 발행하고 전송하는데 사용됩니다. 
- `x/staking` + `x/slashing` : PoS 블록체인을 만드는데 사용됩니다. 

누구나 사용할 수 있는 `x/` 에 이미 있는 모듈들에 더해서 SDK는 당신만의 모듈을 구현할 수 있게 해준다. 그 예시로 [튜토리얼](https://cosmos.network/docs/tutorial/keeper.html)을 여기서 확인할 수 있습니다. 

### 다음에는 Cosmos SDK 의 안전한 모델, [ocap](https://cosmos.network/docs/intro/ocap.html) 에 대해서 더 알아보려고 한다.
