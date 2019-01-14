# Auth 모듈 사양

## 정의

이 문서는 코스모스 SDK의 auth 모듈의 사양을 정의합니다.

SDK 자체는 계정과 트랜잭션에 대한 기준을 정하지 않기 때문에, auth 모듈은 애플리케이션의 기본 트랜잭션과 계정 형태를 정의합니다. 이 모듈에는 모든 트랜잭션(서명, 논스, 기타 값)의 기본적인 진위여부를 확인하고 계정의 키퍼를 알려주는 안티핸들러(ante handler)가 포함돼있습니다. 안티핸들러는 다른 모듈들이 계정을 읽고, 쓰고, 변경할 수 있게 합니다.

해당 모듈은 코스모스 허브에서 이용됩니다.

## 목차

1. **[상태(state)](state.md)**
    1. [계정](state.md#accounts)
        1. [계정 인터페이스](state.md#account-interface)
        1. [베이스 계정(base account)](state.md#baseaccount)
        1. [베스팅 계정(vesting account)](state.md#vestingaccount)
1. **[종류](types.md)**
    1. [StdFee](types.md#stdfee)
    1. [StdSignature](types.md#stdsignature)
    1. [StdTx](types.md#stdtx)
    1. [StdSignDoc](types.md#stdsigndoc)
1. **[키퍼](keepers.md)**
    1. [AccountKeeper](keepers.md#account-keeper)
1. **[핸들러](handlers.md)**
    1. [Ante Handler](handlers.md#ante-handler)
