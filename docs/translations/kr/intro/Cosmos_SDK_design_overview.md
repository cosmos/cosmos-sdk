## Cosmos SDK design overview 

Cosmos SDK 는 텐더민트 상에서 돌아가는 안전한 상태 머신의 개발을 촉진시켜주는 프레임워크입니다. 이 코어에서, SDK 는 Go 언어로 작성된 ABCI 의 boilerplate 구현체입니다. `multiscore` 는 데이터를 지속시키는데 사용되고 `router` 는 거래를 처리하는데 사용됩니다. 

Cosmos SDK 로 구현된 어플리케이션이 `DeliverTx` 를 통해서 텐더민트에게 넘어온 거래를 어떻게 처리하는지 간단하게 설명해보겠습니다 : 

1. 텐더민트 합의 엔진으로 부터 받은 `transaction` 을 deocde 합니다. (기억하십시오. 텐더민트 `[]bytes`의 형태선에서만 처리합니다. 그 내용은 무엇인지 인지하지 않습니다.)
2. `transaction`로 부터 `messages` 를 추출하여 기본적인 온전성을 검사합니다. 
3. 잘 처리될 수 있도록 각 메세지를 적절한 module 에 route 해줍니다. 
4. 상태 변환을 커밋합니다. 

어플리케이션이 당신으로 하여금 거래를 생성하게 할 수도 있습니다, 그들을 encode 하고 해당 텐더민트 엔진에 넘겨 그들을 브로드캐스트합니다. 

### `baseapp`

`baseapp` 은 Cosmos SDK, ABCI 의 boilerplate 구현체입니다. `router` 는 거래를 그들과 관련있는 모듈에게 넘겨주는 역할을 합니다. 당신의 어플리케이션의 메인 파알 `app.go` 은 당신의 `app` 타입을 정의할 것이고 `baseapp` 을 임베트(embed) 할 것입니다. 이 방법으로, 당신의 `app` 타입은 자동적으로 `baseapp` 의 ABCI 메소드 전부를 상속할 것입니다. 그 예시는 SDK 어플리케이션 튜토리얼에서 볼 수 있습니다. 

`baseapp` 의 목표는 is to provide a secure interface between the store and the extensible state machine while defining as little about the state machine as possible (staying true to the ABCI).

`baseapp` 에 대한 추가적인 정보는 여기에서 볼 수 있습니다. 

### Multistore 

Cosmos SDK 는 multiscore 에 지속적인 상태를 제공합니다. multiscore 는 개발자로 하여금 다수의 `KVStores`를 선언할 수 있게 해줍니다. 이 `KVStore` 들은 오직 `[]byte` 타입의 값들만 승인합니다. 그러므로 어떤 종류의 구조체도 저장되기 전에 go-amino 로 정렬화 과정을 거쳐야 합니다. 

The multistore abstraction is used to divide the state in distinct compartments, each managed by its own module. For more on the multistore, click here

### Modules 

Cosmos SDK 의 힘은 그 모듈화에서 온다. SDK 어플리케이션은 상호작용하는 모듈들의 모음을 합쳐서 만들어집니다. 각 모듈은 상태의 subset 을 정의하고 그들 고유의 message/transaction 과정을 가지고 있습니다, SDK 는 각 메세지를 이에 어울리는 모듈에 라우팅하는데 사용됩니다. 

아래에 어플리케이션의 각 풀노드가 유효한 블록을 받았을 때 이를 어떻게 처리하는지를 설명하는 과정이 묘사되어있다. 



각 모듈은 작은 상태머신으로 보일 수도 있다. 개발자들은 모듈에 의해서 다뤄지는 상태의 subset 을 정의할 필요가 있다. , as well as custom message types that modify the state (Note: `messages` are extracted from `transactions` using baseapp). 보통, 각 모듈은 그들만의 `KVStore` multiscore 에 선언하고 그것이 정의하는 상태의 subset 을 유지한다. 대부분의 개발자들은 자신만의 모듈을 개발할 때 제 3의 모듈에 접근해야 한다. Given that the Cosmos-SDK is an open framework, some of the modules may be malicious, which means there is a need for security principles to reason about inter-module interactions. These principles are based on object-capabilities. In practice, this means that instead of having each module keep an access control list for other modules, each module implements special objects called keepers that can be passed to other modules to grant a pre-defined set of capabilities.

SDK modules are defined in the x/ folder of the SDK. Some core modules include:

- `x/auth` : 계정과 서명을 관리하는데 사용됩니다. 
- `x/bank` : 토큰을 발행하고 전송하는데 사용됩니다. 
- `x/staking` + `x/slashing` : PoS 블록체인을 만드는데 사용됩니다. 

누구나 사용할 수 있는 `x/` 에 이미 있는 모듈들에 더해서 SDK는 당신만의 모듈을 구현할 수 있게 해준다. 그 예시로 튜토리얼을 여기서 확인할 수 있습니다. 

### 다음에는 Cosmos SDK 의 안전한 모델, ocap 에 대해서 더 알아보려고 한다.
