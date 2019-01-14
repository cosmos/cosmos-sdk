# REST 서버 시작하기

REST 서버를 가동하기 위해서는 다음과 같은 파라미터 값을 정의해야합니다:


| 파라미터   | 형태      | 기본 값                 | 필수/선택 | 설명                                          |
| ----------- | --------- | ----------------------- | -------- | ---------------------------------------------------- |
| chain-id    | string    | null                    | 필수     | 연결할 체인의 chain-id                 |
| node        | URL       | "tcp://localhost:46657" | 필수     | 연결할 풀노드의 주소     |
| laddr       | URL       | "tcp://localhost:1317"  | 필수     | REST 서버를 가동할 주소         |
| trust-node  | bool      | "false"                 | 필수     | 연결할 풀노드의 신뢰가능 여부 |
| trust-store | DIRECTORY | "$HOME/.lcd"            | 선택    | 체크포인와 검증인 세트를 저장할 디렉토리    |

예를 들어::

```bash
gaiacli advanced rest-server --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node tcp://localhost:26657 \
    --trust-node=false
```

서버는 기본적으로 HTTPS를 확인합니다. 서버가 이용하는 SSL에 다음과 같은 추가 플래그를 설정하실 수 있습니다:

```bash
gaiacli advanced rest-server --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node tcp://localhost:26657 \
    --trust-node=false \
    --certfile=mycert.pem --keyfile=mykey.key
```

만약 인증서 또는 키파일 세트가 제공되지 않을 경우, 자체적인 인증서가 생성되고 관련 지문(fingerprint)이 프린트(print) 됩니다. 만약 안전계층을 비활성화 하고 안전하지 않을 수 있는 HTTP 포트로 연결하시는 것을 원하시는 경우 `--insecure` 플래그를 추가해주세요.

Gaia-Lite RPC에 대한 추가적인 정보를 원하시면 [Swagger 문서](https://cosmos.network/rpc/)를 확인하세요.
