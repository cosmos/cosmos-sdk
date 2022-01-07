package cosmos.auth.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Query defines the gRPC querier service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/auth/v1beta1/query.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class QueryGrpc {

  private QueryGrpc() {}

  public static final String SERVICE_NAME = "cosmos.auth.v1beta1.Query";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest,
      cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse> getAccountsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Accounts",
      requestType = cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest.class,
      responseType = cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest,
      cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse> getAccountsMethod() {
    io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest, cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse> getAccountsMethod;
    if ((getAccountsMethod = QueryGrpc.getAccountsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getAccountsMethod = QueryGrpc.getAccountsMethod) == null) {
          QueryGrpc.getAccountsMethod = getAccountsMethod =
              io.grpc.MethodDescriptor.<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest, cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Accounts"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Accounts"))
              .build();
        }
      }
    }
    return getAccountsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest,
      cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse> getAccountMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Account",
      requestType = cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest.class,
      responseType = cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest,
      cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse> getAccountMethod() {
    io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest, cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse> getAccountMethod;
    if ((getAccountMethod = QueryGrpc.getAccountMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getAccountMethod = QueryGrpc.getAccountMethod) == null) {
          QueryGrpc.getAccountMethod = getAccountMethod =
              io.grpc.MethodDescriptor.<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest, cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Account"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Account"))
              .build();
        }
      }
    }
    return getAccountMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest,
      cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Params",
      requestType = cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest.class,
      responseType = cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest,
      cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod() {
    io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest, cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse> getParamsMethod;
    if ((getParamsMethod = QueryGrpc.getParamsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getParamsMethod = QueryGrpc.getParamsMethod) == null) {
          QueryGrpc.getParamsMethod = getParamsMethod =
              io.grpc.MethodDescriptor.<cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest, cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Params"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Params"))
              .build();
        }
      }
    }
    return getParamsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest,
      cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse> getModuleAccountsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ModuleAccounts",
      requestType = cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest.class,
      responseType = cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest,
      cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse> getModuleAccountsMethod() {
    io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest, cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse> getModuleAccountsMethod;
    if ((getModuleAccountsMethod = QueryGrpc.getModuleAccountsMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getModuleAccountsMethod = QueryGrpc.getModuleAccountsMethod) == null) {
          QueryGrpc.getModuleAccountsMethod = getModuleAccountsMethod =
              io.grpc.MethodDescriptor.<cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest, cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ModuleAccounts"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("ModuleAccounts"))
              .build();
        }
      }
    }
    return getModuleAccountsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest,
      cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse> getBech32PrefixMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Bech32Prefix",
      requestType = cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest.class,
      responseType = cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest,
      cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse> getBech32PrefixMethod() {
    io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest, cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse> getBech32PrefixMethod;
    if ((getBech32PrefixMethod = QueryGrpc.getBech32PrefixMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getBech32PrefixMethod = QueryGrpc.getBech32PrefixMethod) == null) {
          QueryGrpc.getBech32PrefixMethod = getBech32PrefixMethod =
              io.grpc.MethodDescriptor.<cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest, cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Bech32Prefix"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Bech32Prefix"))
              .build();
        }
      }
    }
    return getBech32PrefixMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest,
      cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse> getAddressBytesToStringMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "AddressBytesToString",
      requestType = cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest.class,
      responseType = cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest,
      cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse> getAddressBytesToStringMethod() {
    io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest, cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse> getAddressBytesToStringMethod;
    if ((getAddressBytesToStringMethod = QueryGrpc.getAddressBytesToStringMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getAddressBytesToStringMethod = QueryGrpc.getAddressBytesToStringMethod) == null) {
          QueryGrpc.getAddressBytesToStringMethod = getAddressBytesToStringMethod =
              io.grpc.MethodDescriptor.<cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest, cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "AddressBytesToString"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("AddressBytesToString"))
              .build();
        }
      }
    }
    return getAddressBytesToStringMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest,
      cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse> getAddressStringToBytesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "AddressStringToBytes",
      requestType = cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest.class,
      responseType = cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest,
      cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse> getAddressStringToBytesMethod() {
    io.grpc.MethodDescriptor<cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest, cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse> getAddressStringToBytesMethod;
    if ((getAddressStringToBytesMethod = QueryGrpc.getAddressStringToBytesMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getAddressStringToBytesMethod = QueryGrpc.getAddressStringToBytesMethod) == null) {
          QueryGrpc.getAddressStringToBytesMethod = getAddressStringToBytesMethod =
              io.grpc.MethodDescriptor.<cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest, cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "AddressStringToBytes"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("AddressStringToBytes"))
              .build();
        }
      }
    }
    return getAddressStringToBytesMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static QueryStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryStub>() {
        @java.lang.Override
        public QueryStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryStub(channel, callOptions);
        }
      };
    return QueryStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static QueryBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryBlockingStub>() {
        @java.lang.Override
        public QueryBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryBlockingStub(channel, callOptions);
        }
      };
    return QueryBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static QueryFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryFutureStub>() {
        @java.lang.Override
        public QueryFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryFutureStub(channel, callOptions);
        }
      };
    return QueryFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Query defines the gRPC querier service.
   * </pre>
   */
  public static abstract class QueryImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * Accounts returns all the existing accounts
     * Since: cosmos-sdk 0.43
     * </pre>
     */
    public void accounts(cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getAccountsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Account returns account details based on address.
     * </pre>
     */
    public void account(cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getAccountMethod(), responseObserver);
    }

    /**
     * <pre>
     * Params queries all parameters.
     * </pre>
     */
    public void params(cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getParamsMethod(), responseObserver);
    }

    /**
     * <pre>
     * ModuleAccounts returns all the existing module accounts.
     * </pre>
     */
    public void moduleAccounts(cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getModuleAccountsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Bech32 queries bech32Prefix
     * </pre>
     */
    public void bech32Prefix(cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getBech32PrefixMethod(), responseObserver);
    }

    /**
     * <pre>
     * AddressBytesToString converts Account Address bytes to string
     * </pre>
     */
    public void addressBytesToString(cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getAddressBytesToStringMethod(), responseObserver);
    }

    /**
     * <pre>
     * AddressStringToBytes converts Address string to bytes
     * </pre>
     */
    public void addressStringToBytes(cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getAddressStringToBytesMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getAccountsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest,
                cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse>(
                  this, METHODID_ACCOUNTS)))
          .addMethod(
            getAccountMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest,
                cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse>(
                  this, METHODID_ACCOUNT)))
          .addMethod(
            getParamsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest,
                cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse>(
                  this, METHODID_PARAMS)))
          .addMethod(
            getModuleAccountsMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest,
                cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse>(
                  this, METHODID_MODULE_ACCOUNTS)))
          .addMethod(
            getBech32PrefixMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest,
                cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse>(
                  this, METHODID_BECH32PREFIX)))
          .addMethod(
            getAddressBytesToStringMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest,
                cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse>(
                  this, METHODID_ADDRESS_BYTES_TO_STRING)))
          .addMethod(
            getAddressStringToBytesMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest,
                cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse>(
                  this, METHODID_ADDRESS_STRING_TO_BYTES)))
          .build();
    }
  }

  /**
   * <pre>
   * Query defines the gRPC querier service.
   * </pre>
   */
  public static final class QueryStub extends io.grpc.stub.AbstractAsyncStub<QueryStub> {
    private QueryStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryStub(channel, callOptions);
    }

    /**
     * <pre>
     * Accounts returns all the existing accounts
     * Since: cosmos-sdk 0.43
     * </pre>
     */
    public void accounts(cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getAccountsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Account returns account details based on address.
     * </pre>
     */
    public void account(cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getAccountMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Params queries all parameters.
     * </pre>
     */
    public void params(cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getParamsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * ModuleAccounts returns all the existing module accounts.
     * </pre>
     */
    public void moduleAccounts(cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getModuleAccountsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Bech32 queries bech32Prefix
     * </pre>
     */
    public void bech32Prefix(cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getBech32PrefixMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * AddressBytesToString converts Account Address bytes to string
     * </pre>
     */
    public void addressBytesToString(cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getAddressBytesToStringMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * AddressStringToBytes converts Address string to bytes
     * </pre>
     */
    public void addressStringToBytes(cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest request,
        io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getAddressStringToBytesMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Query defines the gRPC querier service.
   * </pre>
   */
  public static final class QueryBlockingStub extends io.grpc.stub.AbstractBlockingStub<QueryBlockingStub> {
    private QueryBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Accounts returns all the existing accounts
     * Since: cosmos-sdk 0.43
     * </pre>
     */
    public cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse accounts(cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getAccountsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Account returns account details based on address.
     * </pre>
     */
    public cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse account(cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getAccountMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Params queries all parameters.
     * </pre>
     */
    public cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse params(cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getParamsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * ModuleAccounts returns all the existing module accounts.
     * </pre>
     */
    public cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse moduleAccounts(cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getModuleAccountsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Bech32 queries bech32Prefix
     * </pre>
     */
    public cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse bech32Prefix(cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getBech32PrefixMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * AddressBytesToString converts Account Address bytes to string
     * </pre>
     */
    public cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse addressBytesToString(cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getAddressBytesToStringMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * AddressStringToBytes converts Address string to bytes
     * </pre>
     */
    public cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse addressStringToBytes(cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getAddressStringToBytesMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Query defines the gRPC querier service.
   * </pre>
   */
  public static final class QueryFutureStub extends io.grpc.stub.AbstractFutureStub<QueryFutureStub> {
    private QueryFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * Accounts returns all the existing accounts
     * Since: cosmos-sdk 0.43
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse> accounts(
        cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getAccountsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Account returns account details based on address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse> account(
        cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getAccountMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Params queries all parameters.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse> params(
        cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getParamsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * ModuleAccounts returns all the existing module accounts.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse> moduleAccounts(
        cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getModuleAccountsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Bech32 queries bech32Prefix
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse> bech32Prefix(
        cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getBech32PrefixMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * AddressBytesToString converts Account Address bytes to string
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse> addressBytesToString(
        cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getAddressBytesToStringMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * AddressStringToBytes converts Address string to bytes
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse> addressStringToBytes(
        cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getAddressStringToBytesMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_ACCOUNTS = 0;
  private static final int METHODID_ACCOUNT = 1;
  private static final int METHODID_PARAMS = 2;
  private static final int METHODID_MODULE_ACCOUNTS = 3;
  private static final int METHODID_BECH32PREFIX = 4;
  private static final int METHODID_ADDRESS_BYTES_TO_STRING = 5;
  private static final int METHODID_ADDRESS_STRING_TO_BYTES = 6;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final QueryImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(QueryImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_ACCOUNTS:
          serviceImpl.accounts((cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountsResponse>) responseObserver);
          break;
        case METHODID_ACCOUNT:
          serviceImpl.account((cosmos.auth.v1beta1.QueryOuterClass.QueryAccountRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryAccountResponse>) responseObserver);
          break;
        case METHODID_PARAMS:
          serviceImpl.params((cosmos.auth.v1beta1.QueryOuterClass.QueryParamsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryParamsResponse>) responseObserver);
          break;
        case METHODID_MODULE_ACCOUNTS:
          serviceImpl.moduleAccounts((cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.QueryModuleAccountsResponse>) responseObserver);
          break;
        case METHODID_BECH32PREFIX:
          serviceImpl.bech32Prefix((cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.Bech32PrefixResponse>) responseObserver);
          break;
        case METHODID_ADDRESS_BYTES_TO_STRING:
          serviceImpl.addressBytesToString((cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.AddressBytesToStringResponse>) responseObserver);
          break;
        case METHODID_ADDRESS_STRING_TO_BYTES:
          serviceImpl.addressStringToBytes((cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.auth.v1beta1.QueryOuterClass.AddressStringToBytesResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  private static abstract class QueryBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    QueryBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return cosmos.auth.v1beta1.QueryOuterClass.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("Query");
    }
  }

  private static final class QueryFileDescriptorSupplier
      extends QueryBaseDescriptorSupplier {
    QueryFileDescriptorSupplier() {}
  }

  private static final class QueryMethodDescriptorSupplier
      extends QueryBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    QueryMethodDescriptorSupplier(String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (QueryGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new QueryFileDescriptorSupplier())
              .addMethod(getAccountsMethod())
              .addMethod(getAccountMethod())
              .addMethod(getParamsMethod())
              .addMethod(getModuleAccountsMethod())
              .addMethod(getBech32PrefixMethod())
              .addMethod(getAddressBytesToStringMethod())
              .addMethod(getAddressStringToBytesMethod())
              .build();
        }
      }
    }
    return result;
  }
}
