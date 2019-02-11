# 오브젝트-가능성 모델(Object-Capability Model)

## 개요

보안을 검토하기 위해서는 특정 위협 모델(threat model)을 검토하는 것으로 시작하는 것이 좋습니다. 우리가 제시하는 모델은 다음과 같습니다:

> 코스모스-SDK 모듈을 기반으로 만들어진 블록체인 애플리케이션이 활성화된 생태계에는 허점이 있거나 악의적인 모듈이 존재할 수 있다고 전재합니다.

코스모스 SDK는 오브젝트-가능성 시스템의 토대가 됨으로 이런 문제점을 해결할 수 있습니다.

> 오브젝트 가능성 시스템의 구조적 속성은 코드 디자인의 모듈화와 안정적인 캡슐화(encapsulation)를 선호합니다. 
>
> 이런 구조적 속성은 오브젝트-가능 프로그램 또는 운영체제의 보안적 속성의 분석을 가능하게 합니다. 정보 플로우 속성(information flow properties) 같은 일부 속성은 특정 오브젝트의 행동을 결정하는 코드의 지식 또는 분석 없이도 오직 오브젝트 레퍼런스와 연결구조만으로 분석이 가능합니다.
>
> 그렇기 때문에, 악의적인 코드가 포함되있을 확률이 있는 새로운 오브젝트가 소개되더라도 보안적 속성은 지켜질 수 있습니다.
>
> 이런 구조적 속성은 해당 오브젝트를 통치하는 두가지 법칙에 의해 지켜질 수 있습니다:
> 1. 오브젝트 'A'는 'B'에 대한 레퍼런스를 보유하고 있을 경우에만 메시지를 전송할 수 있다.
> 2. 오브젝트 'A'가 'C'에 대한 레퍼런스를 가지고 싶다면 오브젝트 'A'는 'C'에 대한 레퍼런스가 포함된 메시지를 수신해야 한다.
>
> 이 두 법칙에 따르면 특정 오브젝트는 기존에 존재하는 레퍼런스 체인을 통해서만 다른 오브젝트에 대한 레퍼런스를 받을 수 있는 것입니다. 단순하게 설명하면 "연결성이 연결성을 불러온다"는 것입니다.

오브젝트 가능성에 대한 소개는 [이 글(영문)](http://habitatchronicles.com/2017/05/what-are-capabilities/)을 참고하세요.

원칙적으로 말하면, 다음과 같은 문제점 때문에 Golang은 엄격한 오브젝트 가능성을 도입하지 않는다고 볼 수 있습니다:

- 보편적으로 "unsafe" 또는 "os" 같은 원시적 모듈(primitive module)을 임포트할 수 있음.
- 보편적으로 [모듈 vars 오버라이드(https://github.com/golang/go/issues/23161)]가 가능.
- 2개 이상의 goroutine이 illegal interface value를 만들 수 있는 data-race 허점.

문제점 중 첫번째 문제는 임포트 감사(import audit)과 체계적인 디펜던시 버전 컨트롤 시스템(dependency version control system)을 이용하여 문제점을 사전에 찾을 수 있습니다. 다만, 두번째와 세번째 문제는 다서 불편하기는 하지만 노력을 통해서 감사(audit)가 가능합니다.

현재로써는 [Go2에서 오브젝트 가능성 모델이 도입되는](https://github.com/golang/go/issues/23157)것을 기대하는 상태입니다.

## 실전에서의 오브젝트 가능성 모델(Ocaps)

Ocaps의 핵심 아이디어는 특정 일을 수행하는데 필요한 정보만을 공개하는 것입니다.

예를 들어, 다음의 코드는 오브젝트 가능성 원칙을 위해합니다:

```go
type AppAccount struct {...}
var account := &AppAccount{
    Address: pub.Address(),
    Coins: sdk.Coins{sdk.NewInt64Coin("ATM", 100)},
}
var sumValue := externalModule.ComputeSumValue(account)
```

위 코드의 `ComputeSumValue` 메소드는 순수한(pure) 함수를 암시하지만, 포인터 밸류를 받아드리는 것의 암시된 가능성(implied capability)은 해당 밸류를 변경할 수 있는 가능성입니다. 서명 메소드는 그대로 포인터 밸류를 받아들이지 않고 해당 밸류의 사본을 사용하는 것이 바람직합니다.

```go
var sumValue := externalModule.ComputeSumValue(*account)
```

코스모스 SDK에서 이 원칙을 응용한 것을 [gaia 앱](../gaia/app/app.go)에서 확인이 가능합니다.

```go
// register message routes
app.Router().
  AddRoute(bank.RouterKey, bank.NewHandler(app.bankKeeper)).
  AddRoute(staking.RouterKey, staking.NewHandler(app.stakingKeeper)).
  AddRoute(distr.RouterKey, distr.NewHandler(app.distrKeeper)).
  AddRoute(slashing.RouterKey, slashing.NewHandler(app.slashingKeeper)).
  AddRoute(gov.RouterKey, gov.NewHandler(app.govKeeper))
```


