# SDK 애플리케이션 아키텍쳐

## SDK 블록체인을 만드는 구성원

블록체인 애플리케이션은 근본적으로 [결정론적 복제 상태 기계(replicated deterministic state machine)](https://ko.wikipedia.org/wiki/%EC%83%81%ED%83%9C_%EA%B8%B0%EA%B3%84_%EB%B3%B5%EC%A0%9C)입니다. 개발자는 특정 상태 기계의 정의를 하면 [*텐더민트*](https://tendermint.com/docs/introduction/introduction.html)는 해당 상태기계의 네트워크상의 결정론적 복제를 전담하는 구조입니다.

>텐더민트는 블록체인의 *네트워킹*과 *합의* 계층을 처리하는 애플리케이션-불가지론적(application-agnostic) 엔진입니다. 실전 상황으로 비유하면, 텐더민트는 트랜잭션의 바이트의 순서를 정하고 해당 바이트들을 전파하는 역할을 하는 것입니다. 텐더민트 코어는 기존 비잔틴 결함 감내(BFT, Byzantine-Fault-Tolerant) 알고리즘을 응용하여 트랜잭션의 순서에 대한 합의를 이룹니다. 텐더민트에 대한 더 자세한 정보는 [여기](https://tendermint.com/docs/introduction/introduction.html)에서 확인이 가능합니다.

텐더민트는 ABCI(애플리케이션 특화 블록체인 인터페이스, Application-specific BlockChain Interface)라는 인터페이스를 통해 거래를 네트워크에서 애플리케이션으로 전달합니다. 블록체인 노드의 아키텍쳐는 다음과 같은 형태입니다.


```
+---------------------+
|                     |
|      애플리케이션      |
|                     |
+--------+---+--------+
         ^   |
         |   | ABCI
         |   v
+--------+---+--------+
|                     |
|                     |
|        텐더민트       |
|                     |
|                     |
+---------------------+
```

개발자는 특정 ABCI 인터페이스를 따로 개발할 필요가 없습니다. 코스모스 SDK는 [baseapp](#baseapp)을 통해 ABCI의 기본 틀을 제공합니다.

## BaseApp

지속성을 [MultiStore] 기반의 ABCI 앱을 통해 처리하고, 라우터(Router)를 통해 거래 핸들링을 하는 애플리케이션. BaseApp은 상태 머신에 대한 정보 정의를 최소화(ABCI 기존 틀에서 크게 벗어나지 않기 위해서)한 스토어(store)와 확장적 상태 머신(extensible state machine) 간 소통할 수 있는 안전한 인터페이스를 제공합니다.

`baseapp`에 대한 더 많은 정보는 [이 링크](../reference/baseapp.md)에 있습니다

## 모듈

코스모스-SDK의 가장 큰 장점은 모듈성입니다. SDK 블록체인은 다수의 맞춤 설정이 가능하고 상호 호환성이 확보되어있는 모듈들을 기반으로 만들어집니다. 모듈들은 `x/` 폴더에서 확인이 가능합니다.

SDK는 `x/`에 있는 모듈 외에도 본인의 모듈을 직접 개발할 수 있게 합니다. 한마디로 SDK 블록체인은 이미 개발된 모듈을 임포트 하고 그 외 모듈들을 직접 만들어서 완성될 수 있습니다.

몇가지 핵심 모듈들은 다음과 같습니다:

- `x/auth`: 계정과 서명 관리.
- `x/bank`: 토큰 허용과 토큰 이동.
- `x/staking` + `x/slashing`: 지분증명 블록체인을 만들때 이용.

## Basecoin

Basecoin은 SDK 스택으로 만들어진 최초의 완성된 애플리케이션입니다. 완성된 애플리케이션의 핸들러 기능(handler functionality)를 사용하기 위해서는 SDK 코어 모듈의 익스텐션들이 필수입니다.

Basecoin은 `x/auth` 와 `x/bank` 모듈을 이용하여 `BaseApp` 상태 머신을 임플리멘트 합니다. 이 `BaseApp`은 트랜잭션 서명자의 검증과 코인 이동에 대한 정의를 내립니다. 이 외에도 추가적으로 `x/ibc`와 간단한 스테이킹 익스텐션이 도입될 수 있습니다.

Basecoin과 기본적인 `x/` 익스텐션은 모든 트랜잭션과 계정의 시리얼라이제이션(serialization) 과정에서 go-amino를 이용합니다.

## SDK 투토리얼

SDK 애플리케이션을 만드는 방법과 위에 설명된 개념들에 대해 더 자세히 알고 싶으시다면 [SDK 애플리케이션 투토리얼](https://github.com/cosmos/sdk-application-tutorial)을 참고하세요.
