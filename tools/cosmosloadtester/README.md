# cosmosloadtester

`cosmosloadtester` is a load-testing tool built on top of [informalsystems/tm-load-test](https://github.com/informalsystems/tm-load-test).
It uses [an enhanced fork of tm-load-test](https://github.com/orijtech/tm-load-test), which provides significantly more-detailed stats such as
latency percentile breakdowns and detailed graphs of QPS over time.

The tool consists of [a Go server](https://github.com/orijtech/cosmosloadtester/blob/main/cmd/server/main.go) which exposes the loadtest service over HTTP using gRPC-web and a [built-in React UI](ui) for scheduling loadtests and visualizing results. [The gRPC service](proto/orijtech/cosmosloadtester/v1/loadtest_service.proto) can also be interacted with without the UI by using [gRPC](https://grpc.io/), [gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway), or [gRPC-web](https://github.com/grpc/grpc-web).

To leverage this tool, you'll need to [write logic to generate transactions for your message type](https://github.com/orijtech/cosmosloadtester/edit/main/README.md#registering-custom-client-factories).

## Building and running the server

1. Build the UI:
```shell
make ui
```

2. Build the server:
```shell
make server
```

3. Run the server:
```shell
./bin/server --port=8080
```
4. The server should be available at http://localhost:8080


## Registering custom client factories

To use this tool, you will need to write a client factory that generates transactions for the message type(s) you want to load-test.

1. [Create your custom client factory](https://github.com/informalsystems/tm-load-test/tree/main/pkg/loadtest#step-2-create-your-load-testing-client).

      For use as a template, a sample client factory that generates an empty Cosmos transaction can be found under [clients/myabciapp/client.go](clients/myabciapp/client.go):
    https://github.com/orijtech/cosmosloadtester/blob/1d66499b0d56fcbfb1888047a7f0ad1c697b8dbf/clients/myabciapp/client.go#L45-L53

2. Register your factory with a meaningful name in `registerClientFactories` in [cmd/server/main.go](cmd/server/main.go):

    https://github.com/orijtech/cosmosloadtester/blob/1d66499b0d56fcbfb1888047a7f0ad1c697b8dbf/cmd/server/main.go#L115-L124

2. After adding and registering your client factory, make sure to [rebuild the server](https://github.com/orijtech/cosmosloadtester/edit/main/README.md#building-and-running-the-server).
3. Then, you can enter its name under `Client factory` in the UI to use it:

    ![image](https://user-images.githubusercontent.com/6455350/208562755-4f6fbdd1-aebb-447c-9394-fadb73a8a50a.png)



## Images

### Input screen
![](https://user-images.githubusercontent.com/6455350/208561926-e5bbe6de-691b-488f-86b2-0f7794a11022.png)

### Results
![](https://user-images.githubusercontent.com/6455350/208562264-1e9f3b5d-1e94-4b62-a6f3-455654683068.png)

### Sequence diagram
![image](https://user-images.githubusercontent.com/6455350/208564793-e141055d-d0c8-42d5-9576-a2b3ed312e38.png)
