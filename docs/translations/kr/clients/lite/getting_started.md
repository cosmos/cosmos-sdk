# REST 서버 시작하기

REST 서버를 가동하기 위해서는 다음과 같은 파라미터 값을 정의해야 합니다:


| 파라미터   | 형태      | 기본 값                 | 필수/선택 | 설명                                          |
| ----------- | --------- | ----------------------- | -------- | ---------------------------------------------------- |
| chain-id    | string    | null                    | 필수     | 연결할 체인의 chain-id                 |
| node        | URL       | "tcp://localhost:46657" | 필수     | 연결할 풀노드의 주소     |
| laddr       | URL       | "tcp://localhost:1317"  | 필수     | REST 서버를 가동할 주소         |
| trust-node  | bool      | "false"                 | 필수     | 연결할 풀노드의 신뢰가능 여부 |
| trust-store | DIRECTORY | "$HOME/.lcd"            | 선택    | 체크포인트와 검증인 세트를 저장할 디렉터리    |

예를 들어::

```bash
gaiacli rest-server --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node tcp://localhost:26657 \
    --trust-node=false
```

서버는 기본적으로 HTTP를 확인합니다. 보안 계층을 사용하시려면 `--tls` 플래그를 추가해주세요. 기본적으로 자체 서명이 된 인증서가 생성되며, fingerprint가 프린트됩니다. 서버에 특정 SSL 인증서를 사용하기 ㅜ이해서는 `--ssl-certfile`과 `--ssl-keyfile` 플래그를 지정해주세요:

```bash
gaiacli rest-server --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node tcp://localhost:26657 \
    --trust-node=false \
    --ssl-certfile=mycert.pem --ssl-keyfile=mykey.key
```

Gaia-Lite RPC에 대한 추가적인 정보를 원하시면 [Swagger 문서](https://cosmos.network/rpc/)를 확인하세요.
