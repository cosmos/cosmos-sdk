package cosmos.nft.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Query defines the gRPC querier service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/nft/v1beta1/query.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class QueryGrpc {

  private QueryGrpc() {}

  public static final String SERVICE_NAME = "cosmos.nft.v1beta1.Query";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse> getBalanceMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Balance",
      requestType = cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest.class,
      responseType = cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse> getBalanceMethod() {
    io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse> getBalanceMethod;
    if ((getBalanceMethod = QueryGrpc.getBalanceMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getBalanceMethod = QueryGrpc.getBalanceMethod) == null) {
          QueryGrpc.getBalanceMethod = getBalanceMethod =
              io.grpc.MethodDescriptor.<cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Balance"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Balance"))
              .build();
        }
      }
    }
    return getBalanceMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse> getOwnerMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Owner",
      requestType = cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest.class,
      responseType = cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse> getOwnerMethod() {
    io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse> getOwnerMethod;
    if ((getOwnerMethod = QueryGrpc.getOwnerMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getOwnerMethod = QueryGrpc.getOwnerMethod) == null) {
          QueryGrpc.getOwnerMethod = getOwnerMethod =
              io.grpc.MethodDescriptor.<cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Owner"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Owner"))
              .build();
        }
      }
    }
    return getOwnerMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse> getSupplyMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Supply",
      requestType = cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest.class,
      responseType = cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse> getSupplyMethod() {
    io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest, cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse> getSupplyMethod;
    if ((getSupplyMethod = QueryGrpc.getSupplyMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getSupplyMethod = QueryGrpc.getSupplyMethod) == null) {
          QueryGrpc.getSupplyMethod = getSupplyMethod =
              io.grpc.MethodDescriptor.<cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest, cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Supply"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Supply"))
              .build();
        }
      }
    }
    return getSupplyMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse> getNFTsOfClassMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "NFTsOfClass",
      requestType = cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest.class,
      responseType = cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse> getNFTsOfClassMethod() {
    io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse> getNFTsOfClassMethod;
    if ((getNFTsOfClassMethod = QueryGrpc.getNFTsOfClassMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getNFTsOfClassMethod = QueryGrpc.getNFTsOfClassMethod) == null) {
          QueryGrpc.getNFTsOfClassMethod = getNFTsOfClassMethod =
              io.grpc.MethodDescriptor.<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "NFTsOfClass"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("NFTsOfClass"))
              .build();
        }
      }
    }
    return getNFTsOfClassMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse> getNFTMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "NFT",
      requestType = cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest.class,
      responseType = cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse> getNFTMethod() {
    io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse> getNFTMethod;
    if ((getNFTMethod = QueryGrpc.getNFTMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getNFTMethod = QueryGrpc.getNFTMethod) == null) {
          QueryGrpc.getNFTMethod = getNFTMethod =
              io.grpc.MethodDescriptor.<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "NFT"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("NFT"))
              .build();
        }
      }
    }
    return getNFTMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse> getClassMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Class",
      requestType = cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest.class,
      responseType = cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse> getClassMethod() {
    io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse> getClassMethod;
    if ((getClassMethod = QueryGrpc.getClassMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getClassMethod = QueryGrpc.getClassMethod) == null) {
          QueryGrpc.getClassMethod = getClassMethod =
              io.grpc.MethodDescriptor.<cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Class"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Class"))
              .build();
        }
      }
    }
    return getClassMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse> getClassesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Classes",
      requestType = cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest.class,
      responseType = cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest,
      cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse> getClassesMethod() {
    io.grpc.MethodDescriptor<cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse> getClassesMethod;
    if ((getClassesMethod = QueryGrpc.getClassesMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getClassesMethod = QueryGrpc.getClassesMethod) == null) {
          QueryGrpc.getClassesMethod = getClassesMethod =
              io.grpc.MethodDescriptor.<cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest, cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Classes"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Classes"))
              .build();
        }
      }
    }
    return getClassesMethod;
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
     * Balance queries the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721
     * </pre>
     */
    public void balance(cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getBalanceMethod(), responseObserver);
    }

    /**
     * <pre>
     * Owner queries the owner of the NFT based on its class and id, same as ownerOf in ERC721
     * </pre>
     */
    public void owner(cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getOwnerMethod(), responseObserver);
    }

    /**
     * <pre>
     * Supply queries the number of NFTs from the given class, same as totalSupply of ERC721.
     * </pre>
     */
    public void supply(cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSupplyMethod(), responseObserver);
    }

    /**
     * <pre>
     * NFTsOfClass queries all NFTs of a given class or optional owner, similar to tokenByIndex in ERC721Enumerable
     * </pre>
     */
    public void nFTsOfClass(cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getNFTsOfClassMethod(), responseObserver);
    }

    /**
     * <pre>
     * NFT queries an NFT based on its class and id.
     * </pre>
     */
    public void nFT(cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getNFTMethod(), responseObserver);
    }

    /**
     * <pre>
     * Class queries an NFT class based on its id
     * </pre>
     */
    public void class_(cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getClassMethod(), responseObserver);
    }

    /**
     * <pre>
     * Classes queries all NFT classes
     * </pre>
     */
    public void classes(cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getClassesMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getBalanceMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest,
                cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse>(
                  this, METHODID_BALANCE)))
          .addMethod(
            getOwnerMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest,
                cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse>(
                  this, METHODID_OWNER)))
          .addMethod(
            getSupplyMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest,
                cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse>(
                  this, METHODID_SUPPLY)))
          .addMethod(
            getNFTsOfClassMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest,
                cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse>(
                  this, METHODID_NFTS_OF_CLASS)))
          .addMethod(
            getNFTMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest,
                cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse>(
                  this, METHODID_NFT)))
          .addMethod(
            getClassMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest,
                cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse>(
                  this, METHODID_CLASS)))
          .addMethod(
            getClassesMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest,
                cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse>(
                  this, METHODID_CLASSES)))
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
     * Balance queries the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721
     * </pre>
     */
    public void balance(cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getBalanceMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Owner queries the owner of the NFT based on its class and id, same as ownerOf in ERC721
     * </pre>
     */
    public void owner(cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getOwnerMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Supply queries the number of NFTs from the given class, same as totalSupply of ERC721.
     * </pre>
     */
    public void supply(cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSupplyMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * NFTsOfClass queries all NFTs of a given class or optional owner, similar to tokenByIndex in ERC721Enumerable
     * </pre>
     */
    public void nFTsOfClass(cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getNFTsOfClassMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * NFT queries an NFT based on its class and id.
     * </pre>
     */
    public void nFT(cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getNFTMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Class queries an NFT class based on its id
     * </pre>
     */
    public void class_(cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getClassMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Classes queries all NFT classes
     * </pre>
     */
    public void classes(cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest request,
        io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getClassesMethod(), getCallOptions()), request, responseObserver);
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
     * Balance queries the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721
     * </pre>
     */
    public cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse balance(cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getBalanceMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Owner queries the owner of the NFT based on its class and id, same as ownerOf in ERC721
     * </pre>
     */
    public cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse owner(cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getOwnerMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Supply queries the number of NFTs from the given class, same as totalSupply of ERC721.
     * </pre>
     */
    public cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse supply(cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSupplyMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * NFTsOfClass queries all NFTs of a given class or optional owner, similar to tokenByIndex in ERC721Enumerable
     * </pre>
     */
    public cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse nFTsOfClass(cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getNFTsOfClassMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * NFT queries an NFT based on its class and id.
     * </pre>
     */
    public cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse nFT(cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getNFTMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Class queries an NFT class based on its id
     * </pre>
     */
    public cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse class_(cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getClassMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Classes queries all NFT classes
     * </pre>
     */
    public cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse classes(cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getClassesMethod(), getCallOptions(), request);
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
     * Balance queries the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse> balance(
        cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getBalanceMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Owner queries the owner of the NFT based on its class and id, same as ownerOf in ERC721
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse> owner(
        cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getOwnerMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Supply queries the number of NFTs from the given class, same as totalSupply of ERC721.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse> supply(
        cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSupplyMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * NFTsOfClass queries all NFTs of a given class or optional owner, similar to tokenByIndex in ERC721Enumerable
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse> nFTsOfClass(
        cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getNFTsOfClassMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * NFT queries an NFT based on its class and id.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse> nFT(
        cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getNFTMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Class queries an NFT class based on its id
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse> class_(
        cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getClassMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Classes queries all NFT classes
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse> classes(
        cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getClassesMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_BALANCE = 0;
  private static final int METHODID_OWNER = 1;
  private static final int METHODID_SUPPLY = 2;
  private static final int METHODID_NFTS_OF_CLASS = 3;
  private static final int METHODID_NFT = 4;
  private static final int METHODID_CLASS = 5;
  private static final int METHODID_CLASSES = 6;

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
        case METHODID_BALANCE:
          serviceImpl.balance((cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryBalanceResponse>) responseObserver);
          break;
        case METHODID_OWNER:
          serviceImpl.owner((cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryOwnerResponse>) responseObserver);
          break;
        case METHODID_SUPPLY:
          serviceImpl.supply((cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QuerySupplyResponse>) responseObserver);
          break;
        case METHODID_NFTS_OF_CLASS:
          serviceImpl.nFTsOfClass((cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTsOfClassResponse>) responseObserver);
          break;
        case METHODID_NFT:
          serviceImpl.nFT((cosmos.nft.v1beta1.QueryOuterClass.QueryNFTRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryNFTResponse>) responseObserver);
          break;
        case METHODID_CLASS:
          serviceImpl.class_((cosmos.nft.v1beta1.QueryOuterClass.QueryClassRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryClassResponse>) responseObserver);
          break;
        case METHODID_CLASSES:
          serviceImpl.classes((cosmos.nft.v1beta1.QueryOuterClass.QueryClassesRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.nft.v1beta1.QueryOuterClass.QueryClassesResponse>) responseObserver);
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
      return cosmos.nft.v1beta1.QueryOuterClass.getDescriptor();
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
              .addMethod(getBalanceMethod())
              .addMethod(getOwnerMethod())
              .addMethod(getSupplyMethod())
              .addMethod(getNFTsOfClassMethod())
              .addMethod(getNFTMethod())
              .addMethod(getClassMethod())
              .addMethod(getClassesMethod())
              .build();
        }
      }
    }
    return result;
  }
}
