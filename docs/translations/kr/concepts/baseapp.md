# 베이스앱(BaseApp)

베이스앱(BaseApp)은 코스모스 SDK 애플리케이션이 텐더민트 노드와 소통할 수 있게하는 기본적 ABCI 애플리케이션의 기반을 정의합니다.

베이스앱은 다수의 내부 컴포넌트로 구성되어있습니다. 이 중 가장 중요한 컴포넌트는 `CommitMultiStore`와 해당 컴포넌트의 내부 상태(internal state)입니다. 내부 상태는 하위 두 개의 서브스테이트(substate)로 구성되어있으며, 두 서브 스테이트는 각자 트랜잭션 실행 단계의 `CheckTx`와 `DeliverTx` 이용됩니다. 블록 커밋 단계에서는 `DeliverTx`만이 유지됩니다.

베이스앱은 capabilities keys를 통해 스토어에 마운트 되어야 합니다. 핸들러는 키가 부여된 스토어에만 접근할 수 있습니다. 베이스앱은 모든 스토어가 정상적으로 로딩, 캐쉬 그리고 커밋되는 것을 보장합니다. 마운트 된 스토어 중 '메인(`baseApp.MainstoreKey`' 스토어는 최신 블록 헤더 정보를 보관하며, 이를 기반으로 가장 최근 상태(state)를 찾고 불러올 수 있습니다.

베이스앱은 `AnteHandler`와 `MsgHandler` 두 가지의 핸들러 타입을 받습니다. `AnteHandler`는 논스값 확인, 서명 확인, 수수료 지불 가능 잔고 확인 등 모든 모듈에서 이용되는 트랜잭션 기능의 글로벌 유효성 검증(global validity check)을 수행합니다. `MsgHandler`의 경우 상태 변경(full state transition) 기능을 수행합니다.

`CheckTx` 단계에서 상태 변경 기능은 `checkTxState`에만 적용되며, 값비싼 스테이트 변경 작업이 실행되기 전에 리턴을 해야합니다(관련 절차 설계는 각 개발자의 자유입니다). 또한 예상 가스 비용을 리턴해야 합니다.

`DeliverTx` 단계에서는 상태 변경 기능이 블록체인 상태에 적용되며 트랜잭션은 완전하게 실행되어야 합니다.

베이스앱은 핸들러에게 전달되는 컨텍스트(context)를 관리하는 역할을 합니다. 또한 `CheckTx`와 `DeliverTx`에게 할당되는 스토어와 블록헤더 정보를 제공합니다. 베이스앱은 시리얼화 포맷(serialization format)에 불가지론적(agnostic)입니다.

## 트랜잭션 라이프 사이클(Transaction Life Cycle)

트랜잭션은 실행 단계에서 ABCI 스펙 정의에 따라 `CheckTx`와 `DeliverTx`를 거칠 수 있습니다. `CheckTx`는 블록 제안 검증인이 실행하며 모든 풀노드의 텐더민트 멤풀에서 이용됩니다.

`CheckTx`와 `DeliverTx`는 (정의된 경우) 애플리케이션의 안티핸들러(AnteHandler)를 실행합니다. 여기서 안티핸들러는 메시지-전(pre-message) 검증을 수행하여 계정, 서명 검증, 수수료 삭감, 수수료 수금, 시퀀스 번호 추가 등의 기능을 수행합니다.

### CheckTx

`CheckTx`의 실행 단계에서는 안티핸들러(AnteHandler)만이 실행됩니다.

안티핸들러로 발생하는 상태 변경은 안티핸들러가 실패하여 중단되는 경우를 제외하고는 check-tx 상태의 `CheckTx`의 추후 콜(call)까지 지속됩니다.

### DeliverTx

`DeliverTx`의 실행 단계에서는 안티핸들러(AnteHandler)와 핸들러(Handler)가 함께 실행됩니다.

`DeliverTx`의 트랜잭션 실행 절차는 `CheckTx`와 유사하게 진행됩니다. 다만, 안티핸들러로 발생하는 상태 변경은 핸들러의 프로세싱 로직이 실패하여도 지속됩니다.

악의적 제안자(proposer)가 안티핸들러를 통과하지 못하는 트랜잭션을 발생할 수 있는 경우가 발생할 수도 있습니다. 이 경우에는 악의적 트랜잭션의 상태 변경은 버려집니다(무시됩니다).

## 다른 ABCI 메시지

`CheckTx`와 `DeliverTx` 외에도 베이스앱은 다음과 같은 ABCI 메시지를 처리합니다.

### Info
TODO complete description (추후 업데이트 예정)

### SetOption
TODO complete description (추후 업데이트 예정)

### Query
TODO complete description (추후 업데이트 예정)

### InitChain
TODO complete description (추후 업데이트 예정)

체인 시동(chain initialization) 단계에서 `InitChain`은 `CommitMultiStore`에 직접적으로 할당되어 있는 시동 로직을 실행합니다. check state와 deliver state는 정의된 ChainID로 시작됩니다.

참고할 것은 InitChain 이후에 커밋을 실행하지 않습니다. 그렇기 때문에 블록 1의 BeginBlock은 InitChain이 시작한대로 deliver state에서 시작됩니다.

### BeginBlock
TODO complete description (추후 업데이트 예정)

### EndBlock
TODO complete description (추후 업데이트 예정)

### Commit
TODO complete description (추후 업데이트 예정)


## 가스 관리(Gas Management)

### 가스: InitChain

InitChain 실행 단계에서 블록 가스 미터는 제네시스 트랜잭션을 처리하기 위하여 무한대 가스 수량을 기준으로 시작됩니다.

또한, InitChain의 리퀘스트 메시지에는 genesis.json 파일이 정의하는 ConsensusParams가 포함되어있습니다.


### 가스: BeginBlock

블록 가스 미터는 BeginBlock의 deliver state에서 리셋됩니다. 만약 베이스앱에서 최대 블록 가스가 설정되어있지 않은 경우, 가스 미터는 무한으로 설정됩니다. 최대 블록 가스가 설정되었을 경우, 가스 미터는 `ConsensusParam.BlockSize.MaxGas`를 통해 설정됩니다.

### 가스: DeliverTx

특정 트랜잭션이 실행되기 전, `BlockGasMeter`를 우선 확인하여 남은 가스가 있는지 확인합니다. 만약 남은 가스가 없다면 `DeliverTx`는 즉시 에러를 리턴합니다.

트랜잭션이 처리된 후, 사용된 가스는 (설정된 가스 리밋에 따라) `BlockGasMeter`에서 차감됩니다. 만약 잔류 가스가 가스 미터의 한도를 초과할 경우, `DeliverTx`는 에러를 리턴하고 해당 트랜잭션은 커밋되지 않습니다.

