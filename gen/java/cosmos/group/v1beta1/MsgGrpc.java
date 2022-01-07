package cosmos.group.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Msg is the cosmos.group.v1beta1 Msg service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/group/v1beta1/tx.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class MsgGrpc {

  private MsgGrpc() {}

  public static final String SERVICE_NAME = "cosmos.group.v1beta1.Msg";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgCreateGroup,
      cosmos.group.v1beta1.Tx.MsgCreateGroupResponse> getCreateGroupMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateGroup",
      requestType = cosmos.group.v1beta1.Tx.MsgCreateGroup.class,
      responseType = cosmos.group.v1beta1.Tx.MsgCreateGroupResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgCreateGroup,
      cosmos.group.v1beta1.Tx.MsgCreateGroupResponse> getCreateGroupMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgCreateGroup, cosmos.group.v1beta1.Tx.MsgCreateGroupResponse> getCreateGroupMethod;
    if ((getCreateGroupMethod = MsgGrpc.getCreateGroupMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getCreateGroupMethod = MsgGrpc.getCreateGroupMethod) == null) {
          MsgGrpc.getCreateGroupMethod = getCreateGroupMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgCreateGroup, cosmos.group.v1beta1.Tx.MsgCreateGroupResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateGroup"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgCreateGroup.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgCreateGroupResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("CreateGroup"))
              .build();
        }
      }
    }
    return getCreateGroupMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse> getUpdateGroupMembersMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateGroupMembers",
      requestType = cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers.class,
      responseType = cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse> getUpdateGroupMembersMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers, cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse> getUpdateGroupMembersMethod;
    if ((getUpdateGroupMembersMethod = MsgGrpc.getUpdateGroupMembersMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getUpdateGroupMembersMethod = MsgGrpc.getUpdateGroupMembersMethod) == null) {
          MsgGrpc.getUpdateGroupMembersMethod = getUpdateGroupMembersMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers, cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateGroupMembers"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("UpdateGroupMembers"))
              .build();
        }
      }
    }
    return getUpdateGroupMembersMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse> getUpdateGroupAdminMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateGroupAdmin",
      requestType = cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin.class,
      responseType = cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse> getUpdateGroupAdminMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin, cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse> getUpdateGroupAdminMethod;
    if ((getUpdateGroupAdminMethod = MsgGrpc.getUpdateGroupAdminMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getUpdateGroupAdminMethod = MsgGrpc.getUpdateGroupAdminMethod) == null) {
          MsgGrpc.getUpdateGroupAdminMethod = getUpdateGroupAdminMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin, cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateGroupAdmin"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("UpdateGroupAdmin"))
              .build();
        }
      }
    }
    return getUpdateGroupAdminMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse> getUpdateGroupMetadataMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateGroupMetadata",
      requestType = cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata.class,
      responseType = cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse> getUpdateGroupMetadataMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata, cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse> getUpdateGroupMetadataMethod;
    if ((getUpdateGroupMetadataMethod = MsgGrpc.getUpdateGroupMetadataMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getUpdateGroupMetadataMethod = MsgGrpc.getUpdateGroupMetadataMethod) == null) {
          MsgGrpc.getUpdateGroupMetadataMethod = getUpdateGroupMetadataMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata, cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateGroupMetadata"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("UpdateGroupMetadata"))
              .build();
        }
      }
    }
    return getUpdateGroupMetadataMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy,
      cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse> getCreateGroupPolicyMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateGroupPolicy",
      requestType = cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy.class,
      responseType = cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy,
      cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse> getCreateGroupPolicyMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy, cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse> getCreateGroupPolicyMethod;
    if ((getCreateGroupPolicyMethod = MsgGrpc.getCreateGroupPolicyMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getCreateGroupPolicyMethod = MsgGrpc.getCreateGroupPolicyMethod) == null) {
          MsgGrpc.getCreateGroupPolicyMethod = getCreateGroupPolicyMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy, cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateGroupPolicy"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("CreateGroupPolicy"))
              .build();
        }
      }
    }
    return getCreateGroupPolicyMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse> getUpdateGroupPolicyAdminMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateGroupPolicyAdmin",
      requestType = cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin.class,
      responseType = cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse> getUpdateGroupPolicyAdminMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin, cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse> getUpdateGroupPolicyAdminMethod;
    if ((getUpdateGroupPolicyAdminMethod = MsgGrpc.getUpdateGroupPolicyAdminMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getUpdateGroupPolicyAdminMethod = MsgGrpc.getUpdateGroupPolicyAdminMethod) == null) {
          MsgGrpc.getUpdateGroupPolicyAdminMethod = getUpdateGroupPolicyAdminMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin, cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateGroupPolicyAdmin"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("UpdateGroupPolicyAdmin"))
              .build();
        }
      }
    }
    return getUpdateGroupPolicyAdminMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse> getUpdateGroupPolicyDecisionPolicyMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateGroupPolicyDecisionPolicy",
      requestType = cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy.class,
      responseType = cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse> getUpdateGroupPolicyDecisionPolicyMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy, cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse> getUpdateGroupPolicyDecisionPolicyMethod;
    if ((getUpdateGroupPolicyDecisionPolicyMethod = MsgGrpc.getUpdateGroupPolicyDecisionPolicyMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getUpdateGroupPolicyDecisionPolicyMethod = MsgGrpc.getUpdateGroupPolicyDecisionPolicyMethod) == null) {
          MsgGrpc.getUpdateGroupPolicyDecisionPolicyMethod = getUpdateGroupPolicyDecisionPolicyMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy, cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateGroupPolicyDecisionPolicy"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("UpdateGroupPolicyDecisionPolicy"))
              .build();
        }
      }
    }
    return getUpdateGroupPolicyDecisionPolicyMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse> getUpdateGroupPolicyMetadataMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateGroupPolicyMetadata",
      requestType = cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata.class,
      responseType = cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata,
      cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse> getUpdateGroupPolicyMetadataMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata, cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse> getUpdateGroupPolicyMetadataMethod;
    if ((getUpdateGroupPolicyMetadataMethod = MsgGrpc.getUpdateGroupPolicyMetadataMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getUpdateGroupPolicyMetadataMethod = MsgGrpc.getUpdateGroupPolicyMetadataMethod) == null) {
          MsgGrpc.getUpdateGroupPolicyMetadataMethod = getUpdateGroupPolicyMetadataMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata, cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateGroupPolicyMetadata"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("UpdateGroupPolicyMetadata"))
              .build();
        }
      }
    }
    return getUpdateGroupPolicyMetadataMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgCreateProposal,
      cosmos.group.v1beta1.Tx.MsgCreateProposalResponse> getCreateProposalMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateProposal",
      requestType = cosmos.group.v1beta1.Tx.MsgCreateProposal.class,
      responseType = cosmos.group.v1beta1.Tx.MsgCreateProposalResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgCreateProposal,
      cosmos.group.v1beta1.Tx.MsgCreateProposalResponse> getCreateProposalMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgCreateProposal, cosmos.group.v1beta1.Tx.MsgCreateProposalResponse> getCreateProposalMethod;
    if ((getCreateProposalMethod = MsgGrpc.getCreateProposalMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getCreateProposalMethod = MsgGrpc.getCreateProposalMethod) == null) {
          MsgGrpc.getCreateProposalMethod = getCreateProposalMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgCreateProposal, cosmos.group.v1beta1.Tx.MsgCreateProposalResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateProposal"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgCreateProposal.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgCreateProposalResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("CreateProposal"))
              .build();
        }
      }
    }
    return getCreateProposalMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgVote,
      cosmos.group.v1beta1.Tx.MsgVoteResponse> getVoteMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Vote",
      requestType = cosmos.group.v1beta1.Tx.MsgVote.class,
      responseType = cosmos.group.v1beta1.Tx.MsgVoteResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgVote,
      cosmos.group.v1beta1.Tx.MsgVoteResponse> getVoteMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgVote, cosmos.group.v1beta1.Tx.MsgVoteResponse> getVoteMethod;
    if ((getVoteMethod = MsgGrpc.getVoteMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getVoteMethod = MsgGrpc.getVoteMethod) == null) {
          MsgGrpc.getVoteMethod = getVoteMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgVote, cosmos.group.v1beta1.Tx.MsgVoteResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Vote"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgVote.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgVoteResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("Vote"))
              .build();
        }
      }
    }
    return getVoteMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgExec,
      cosmos.group.v1beta1.Tx.MsgExecResponse> getExecMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Exec",
      requestType = cosmos.group.v1beta1.Tx.MsgExec.class,
      responseType = cosmos.group.v1beta1.Tx.MsgExecResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgExec,
      cosmos.group.v1beta1.Tx.MsgExecResponse> getExecMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.Tx.MsgExec, cosmos.group.v1beta1.Tx.MsgExecResponse> getExecMethod;
    if ((getExecMethod = MsgGrpc.getExecMethod) == null) {
      synchronized (MsgGrpc.class) {
        if ((getExecMethod = MsgGrpc.getExecMethod) == null) {
          MsgGrpc.getExecMethod = getExecMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.Tx.MsgExec, cosmos.group.v1beta1.Tx.MsgExecResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Exec"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgExec.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.Tx.MsgExecResponse.getDefaultInstance()))
              .setSchemaDescriptor(new MsgMethodDescriptorSupplier("Exec"))
              .build();
        }
      }
    }
    return getExecMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static MsgStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<MsgStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<MsgStub>() {
        @java.lang.Override
        public MsgStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new MsgStub(channel, callOptions);
        }
      };
    return MsgStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static MsgBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<MsgBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<MsgBlockingStub>() {
        @java.lang.Override
        public MsgBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new MsgBlockingStub(channel, callOptions);
        }
      };
    return MsgBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static MsgFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<MsgFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<MsgFutureStub>() {
        @java.lang.Override
        public MsgFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new MsgFutureStub(channel, callOptions);
        }
      };
    return MsgFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Msg is the cosmos.group.v1beta1 Msg service.
   * </pre>
   */
  public static abstract class MsgImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * CreateGroup creates a new group with an admin account address, a list of members and some optional metadata.
     * </pre>
     */
    public void createGroup(cosmos.group.v1beta1.Tx.MsgCreateGroup request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgCreateGroupResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateGroupMethod(), responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupMembers updates the group members with given group id and admin address.
     * </pre>
     */
    public void updateGroupMembers(cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateGroupMembersMethod(), responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupAdmin updates the group admin with given group id and previous admin address.
     * </pre>
     */
    public void updateGroupAdmin(cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateGroupAdminMethod(), responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupMetadata updates the group metadata with given group id and admin address.
     * </pre>
     */
    public void updateGroupMetadata(cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateGroupMetadataMethod(), responseObserver);
    }

    /**
     * <pre>
     * CreateGroupPolicy creates a new group policy using given DecisionPolicy.
     * </pre>
     */
    public void createGroupPolicy(cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateGroupPolicyMethod(), responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupPolicyAdmin updates a group policy admin.
     * </pre>
     */
    public void updateGroupPolicyAdmin(cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateGroupPolicyAdminMethod(), responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupPolicyDecisionPolicy allows a group policy's decision policy to be updated.
     * </pre>
     */
    public void updateGroupPolicyDecisionPolicy(cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateGroupPolicyDecisionPolicyMethod(), responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupPolicyMetadata updates a group policy metadata.
     * </pre>
     */
    public void updateGroupPolicyMetadata(cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateGroupPolicyMetadataMethod(), responseObserver);
    }

    /**
     * <pre>
     * CreateProposal submits a new proposal.
     * </pre>
     */
    public void createProposal(cosmos.group.v1beta1.Tx.MsgCreateProposal request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgCreateProposalResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateProposalMethod(), responseObserver);
    }

    /**
     * <pre>
     * Vote allows a voter to vote on a proposal.
     * </pre>
     */
    public void vote(cosmos.group.v1beta1.Tx.MsgVote request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgVoteResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getVoteMethod(), responseObserver);
    }

    /**
     * <pre>
     * Exec executes a proposal.
     * </pre>
     */
    public void exec(cosmos.group.v1beta1.Tx.MsgExec request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgExecResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getExecMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getCreateGroupMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgCreateGroup,
                cosmos.group.v1beta1.Tx.MsgCreateGroupResponse>(
                  this, METHODID_CREATE_GROUP)))
          .addMethod(
            getUpdateGroupMembersMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers,
                cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse>(
                  this, METHODID_UPDATE_GROUP_MEMBERS)))
          .addMethod(
            getUpdateGroupAdminMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin,
                cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse>(
                  this, METHODID_UPDATE_GROUP_ADMIN)))
          .addMethod(
            getUpdateGroupMetadataMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata,
                cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse>(
                  this, METHODID_UPDATE_GROUP_METADATA)))
          .addMethod(
            getCreateGroupPolicyMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy,
                cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse>(
                  this, METHODID_CREATE_GROUP_POLICY)))
          .addMethod(
            getUpdateGroupPolicyAdminMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin,
                cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse>(
                  this, METHODID_UPDATE_GROUP_POLICY_ADMIN)))
          .addMethod(
            getUpdateGroupPolicyDecisionPolicyMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy,
                cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse>(
                  this, METHODID_UPDATE_GROUP_POLICY_DECISION_POLICY)))
          .addMethod(
            getUpdateGroupPolicyMetadataMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata,
                cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse>(
                  this, METHODID_UPDATE_GROUP_POLICY_METADATA)))
          .addMethod(
            getCreateProposalMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgCreateProposal,
                cosmos.group.v1beta1.Tx.MsgCreateProposalResponse>(
                  this, METHODID_CREATE_PROPOSAL)))
          .addMethod(
            getVoteMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgVote,
                cosmos.group.v1beta1.Tx.MsgVoteResponse>(
                  this, METHODID_VOTE)))
          .addMethod(
            getExecMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.Tx.MsgExec,
                cosmos.group.v1beta1.Tx.MsgExecResponse>(
                  this, METHODID_EXEC)))
          .build();
    }
  }

  /**
   * <pre>
   * Msg is the cosmos.group.v1beta1 Msg service.
   * </pre>
   */
  public static final class MsgStub extends io.grpc.stub.AbstractAsyncStub<MsgStub> {
    private MsgStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected MsgStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new MsgStub(channel, callOptions);
    }

    /**
     * <pre>
     * CreateGroup creates a new group with an admin account address, a list of members and some optional metadata.
     * </pre>
     */
    public void createGroup(cosmos.group.v1beta1.Tx.MsgCreateGroup request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgCreateGroupResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateGroupMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupMembers updates the group members with given group id and admin address.
     * </pre>
     */
    public void updateGroupMembers(cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateGroupMembersMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupAdmin updates the group admin with given group id and previous admin address.
     * </pre>
     */
    public void updateGroupAdmin(cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateGroupAdminMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupMetadata updates the group metadata with given group id and admin address.
     * </pre>
     */
    public void updateGroupMetadata(cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateGroupMetadataMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * CreateGroupPolicy creates a new group policy using given DecisionPolicy.
     * </pre>
     */
    public void createGroupPolicy(cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateGroupPolicyMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupPolicyAdmin updates a group policy admin.
     * </pre>
     */
    public void updateGroupPolicyAdmin(cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateGroupPolicyAdminMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupPolicyDecisionPolicy allows a group policy's decision policy to be updated.
     * </pre>
     */
    public void updateGroupPolicyDecisionPolicy(cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateGroupPolicyDecisionPolicyMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * UpdateGroupPolicyMetadata updates a group policy metadata.
     * </pre>
     */
    public void updateGroupPolicyMetadata(cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateGroupPolicyMetadataMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * CreateProposal submits a new proposal.
     * </pre>
     */
    public void createProposal(cosmos.group.v1beta1.Tx.MsgCreateProposal request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgCreateProposalResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateProposalMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Vote allows a voter to vote on a proposal.
     * </pre>
     */
    public void vote(cosmos.group.v1beta1.Tx.MsgVote request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgVoteResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getVoteMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Exec executes a proposal.
     * </pre>
     */
    public void exec(cosmos.group.v1beta1.Tx.MsgExec request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgExecResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getExecMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Msg is the cosmos.group.v1beta1 Msg service.
   * </pre>
   */
  public static final class MsgBlockingStub extends io.grpc.stub.AbstractBlockingStub<MsgBlockingStub> {
    private MsgBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected MsgBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new MsgBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * CreateGroup creates a new group with an admin account address, a list of members and some optional metadata.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgCreateGroupResponse createGroup(cosmos.group.v1beta1.Tx.MsgCreateGroup request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateGroupMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * UpdateGroupMembers updates the group members with given group id and admin address.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse updateGroupMembers(cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateGroupMembersMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * UpdateGroupAdmin updates the group admin with given group id and previous admin address.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse updateGroupAdmin(cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateGroupAdminMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * UpdateGroupMetadata updates the group metadata with given group id and admin address.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse updateGroupMetadata(cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateGroupMetadataMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * CreateGroupPolicy creates a new group policy using given DecisionPolicy.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse createGroupPolicy(cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateGroupPolicyMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * UpdateGroupPolicyAdmin updates a group policy admin.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse updateGroupPolicyAdmin(cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateGroupPolicyAdminMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * UpdateGroupPolicyDecisionPolicy allows a group policy's decision policy to be updated.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse updateGroupPolicyDecisionPolicy(cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateGroupPolicyDecisionPolicyMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * UpdateGroupPolicyMetadata updates a group policy metadata.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse updateGroupPolicyMetadata(cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateGroupPolicyMetadataMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * CreateProposal submits a new proposal.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgCreateProposalResponse createProposal(cosmos.group.v1beta1.Tx.MsgCreateProposal request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateProposalMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Vote allows a voter to vote on a proposal.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgVoteResponse vote(cosmos.group.v1beta1.Tx.MsgVote request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getVoteMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Exec executes a proposal.
     * </pre>
     */
    public cosmos.group.v1beta1.Tx.MsgExecResponse exec(cosmos.group.v1beta1.Tx.MsgExec request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getExecMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Msg is the cosmos.group.v1beta1 Msg service.
   * </pre>
   */
  public static final class MsgFutureStub extends io.grpc.stub.AbstractFutureStub<MsgFutureStub> {
    private MsgFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected MsgFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new MsgFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * CreateGroup creates a new group with an admin account address, a list of members and some optional metadata.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgCreateGroupResponse> createGroup(
        cosmos.group.v1beta1.Tx.MsgCreateGroup request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateGroupMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * UpdateGroupMembers updates the group members with given group id and admin address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse> updateGroupMembers(
        cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateGroupMembersMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * UpdateGroupAdmin updates the group admin with given group id and previous admin address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse> updateGroupAdmin(
        cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateGroupAdminMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * UpdateGroupMetadata updates the group metadata with given group id and admin address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse> updateGroupMetadata(
        cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateGroupMetadataMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * CreateGroupPolicy creates a new group policy using given DecisionPolicy.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse> createGroupPolicy(
        cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateGroupPolicyMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * UpdateGroupPolicyAdmin updates a group policy admin.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse> updateGroupPolicyAdmin(
        cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateGroupPolicyAdminMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * UpdateGroupPolicyDecisionPolicy allows a group policy's decision policy to be updated.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse> updateGroupPolicyDecisionPolicy(
        cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateGroupPolicyDecisionPolicyMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * UpdateGroupPolicyMetadata updates a group policy metadata.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse> updateGroupPolicyMetadata(
        cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateGroupPolicyMetadataMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * CreateProposal submits a new proposal.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgCreateProposalResponse> createProposal(
        cosmos.group.v1beta1.Tx.MsgCreateProposal request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateProposalMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Vote allows a voter to vote on a proposal.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgVoteResponse> vote(
        cosmos.group.v1beta1.Tx.MsgVote request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getVoteMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Exec executes a proposal.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.Tx.MsgExecResponse> exec(
        cosmos.group.v1beta1.Tx.MsgExec request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getExecMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_CREATE_GROUP = 0;
  private static final int METHODID_UPDATE_GROUP_MEMBERS = 1;
  private static final int METHODID_UPDATE_GROUP_ADMIN = 2;
  private static final int METHODID_UPDATE_GROUP_METADATA = 3;
  private static final int METHODID_CREATE_GROUP_POLICY = 4;
  private static final int METHODID_UPDATE_GROUP_POLICY_ADMIN = 5;
  private static final int METHODID_UPDATE_GROUP_POLICY_DECISION_POLICY = 6;
  private static final int METHODID_UPDATE_GROUP_POLICY_METADATA = 7;
  private static final int METHODID_CREATE_PROPOSAL = 8;
  private static final int METHODID_VOTE = 9;
  private static final int METHODID_EXEC = 10;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final MsgImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(MsgImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_CREATE_GROUP:
          serviceImpl.createGroup((cosmos.group.v1beta1.Tx.MsgCreateGroup) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgCreateGroupResponse>) responseObserver);
          break;
        case METHODID_UPDATE_GROUP_MEMBERS:
          serviceImpl.updateGroupMembers((cosmos.group.v1beta1.Tx.MsgUpdateGroupMembers) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupMembersResponse>) responseObserver);
          break;
        case METHODID_UPDATE_GROUP_ADMIN:
          serviceImpl.updateGroupAdmin((cosmos.group.v1beta1.Tx.MsgUpdateGroupAdmin) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupAdminResponse>) responseObserver);
          break;
        case METHODID_UPDATE_GROUP_METADATA:
          serviceImpl.updateGroupMetadata((cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadata) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupMetadataResponse>) responseObserver);
          break;
        case METHODID_CREATE_GROUP_POLICY:
          serviceImpl.createGroupPolicy((cosmos.group.v1beta1.Tx.MsgCreateGroupPolicy) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgCreateGroupPolicyResponse>) responseObserver);
          break;
        case METHODID_UPDATE_GROUP_POLICY_ADMIN:
          serviceImpl.updateGroupPolicyAdmin((cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdmin) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyAdminResponse>) responseObserver);
          break;
        case METHODID_UPDATE_GROUP_POLICY_DECISION_POLICY:
          serviceImpl.updateGroupPolicyDecisionPolicy((cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicy) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyDecisionPolicyResponse>) responseObserver);
          break;
        case METHODID_UPDATE_GROUP_POLICY_METADATA:
          serviceImpl.updateGroupPolicyMetadata((cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadata) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgUpdateGroupPolicyMetadataResponse>) responseObserver);
          break;
        case METHODID_CREATE_PROPOSAL:
          serviceImpl.createProposal((cosmos.group.v1beta1.Tx.MsgCreateProposal) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgCreateProposalResponse>) responseObserver);
          break;
        case METHODID_VOTE:
          serviceImpl.vote((cosmos.group.v1beta1.Tx.MsgVote) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgVoteResponse>) responseObserver);
          break;
        case METHODID_EXEC:
          serviceImpl.exec((cosmos.group.v1beta1.Tx.MsgExec) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.Tx.MsgExecResponse>) responseObserver);
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

  private static abstract class MsgBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    MsgBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return cosmos.group.v1beta1.Tx.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("Msg");
    }
  }

  private static final class MsgFileDescriptorSupplier
      extends MsgBaseDescriptorSupplier {
    MsgFileDescriptorSupplier() {}
  }

  private static final class MsgMethodDescriptorSupplier
      extends MsgBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    MsgMethodDescriptorSupplier(String methodName) {
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
      synchronized (MsgGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new MsgFileDescriptorSupplier())
              .addMethod(getCreateGroupMethod())
              .addMethod(getUpdateGroupMembersMethod())
              .addMethod(getUpdateGroupAdminMethod())
              .addMethod(getUpdateGroupMetadataMethod())
              .addMethod(getCreateGroupPolicyMethod())
              .addMethod(getUpdateGroupPolicyAdminMethod())
              .addMethod(getUpdateGroupPolicyDecisionPolicyMethod())
              .addMethod(getUpdateGroupPolicyMetadataMethod())
              .addMethod(getCreateProposalMethod())
              .addMethod(getVoteMethod())
              .addMethod(getExecMethod())
              .build();
        }
      }
    }
    return result;
  }
}
