package cosmos.group.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Query is the cosmos.group.v1beta1 Query service.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.43.1)",
    comments = "Source: cosmos/group/v1beta1/query.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class QueryGrpc {

  private QueryGrpc() {}

  public static final String SERVICE_NAME = "cosmos.group.v1beta1.Query";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse> getGroupInfoMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GroupInfo",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse> getGroupInfoMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse> getGroupInfoMethod;
    if ((getGroupInfoMethod = QueryGrpc.getGroupInfoMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getGroupInfoMethod = QueryGrpc.getGroupInfoMethod) == null) {
          QueryGrpc.getGroupInfoMethod = getGroupInfoMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GroupInfo"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("GroupInfo"))
              .build();
        }
      }
    }
    return getGroupInfoMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse> getGroupPolicyInfoMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GroupPolicyInfo",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse> getGroupPolicyInfoMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse> getGroupPolicyInfoMethod;
    if ((getGroupPolicyInfoMethod = QueryGrpc.getGroupPolicyInfoMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getGroupPolicyInfoMethod = QueryGrpc.getGroupPolicyInfoMethod) == null) {
          QueryGrpc.getGroupPolicyInfoMethod = getGroupPolicyInfoMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GroupPolicyInfo"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("GroupPolicyInfo"))
              .build();
        }
      }
    }
    return getGroupPolicyInfoMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse> getGroupMembersMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GroupMembers",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse> getGroupMembersMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse> getGroupMembersMethod;
    if ((getGroupMembersMethod = QueryGrpc.getGroupMembersMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getGroupMembersMethod = QueryGrpc.getGroupMembersMethod) == null) {
          QueryGrpc.getGroupMembersMethod = getGroupMembersMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GroupMembers"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("GroupMembers"))
              .build();
        }
      }
    }
    return getGroupMembersMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse> getGroupsByAdminMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GroupsByAdmin",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse> getGroupsByAdminMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse> getGroupsByAdminMethod;
    if ((getGroupsByAdminMethod = QueryGrpc.getGroupsByAdminMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getGroupsByAdminMethod = QueryGrpc.getGroupsByAdminMethod) == null) {
          QueryGrpc.getGroupsByAdminMethod = getGroupsByAdminMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GroupsByAdmin"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("GroupsByAdmin"))
              .build();
        }
      }
    }
    return getGroupsByAdminMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse> getGroupPoliciesByGroupMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GroupPoliciesByGroup",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse> getGroupPoliciesByGroupMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse> getGroupPoliciesByGroupMethod;
    if ((getGroupPoliciesByGroupMethod = QueryGrpc.getGroupPoliciesByGroupMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getGroupPoliciesByGroupMethod = QueryGrpc.getGroupPoliciesByGroupMethod) == null) {
          QueryGrpc.getGroupPoliciesByGroupMethod = getGroupPoliciesByGroupMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GroupPoliciesByGroup"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("GroupPoliciesByGroup"))
              .build();
        }
      }
    }
    return getGroupPoliciesByGroupMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse> getGroupPoliciesByAdminMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GroupPoliciesByAdmin",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse> getGroupPoliciesByAdminMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse> getGroupPoliciesByAdminMethod;
    if ((getGroupPoliciesByAdminMethod = QueryGrpc.getGroupPoliciesByAdminMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getGroupPoliciesByAdminMethod = QueryGrpc.getGroupPoliciesByAdminMethod) == null) {
          QueryGrpc.getGroupPoliciesByAdminMethod = getGroupPoliciesByAdminMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GroupPoliciesByAdmin"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("GroupPoliciesByAdmin"))
              .build();
        }
      }
    }
    return getGroupPoliciesByAdminMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse> getProposalMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Proposal",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse> getProposalMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest, cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse> getProposalMethod;
    if ((getProposalMethod = QueryGrpc.getProposalMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getProposalMethod = QueryGrpc.getProposalMethod) == null) {
          QueryGrpc.getProposalMethod = getProposalMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest, cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Proposal"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("Proposal"))
              .build();
        }
      }
    }
    return getProposalMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse> getProposalsByGroupPolicyMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ProposalsByGroupPolicy",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse> getProposalsByGroupPolicyMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest, cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse> getProposalsByGroupPolicyMethod;
    if ((getProposalsByGroupPolicyMethod = QueryGrpc.getProposalsByGroupPolicyMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getProposalsByGroupPolicyMethod = QueryGrpc.getProposalsByGroupPolicyMethod) == null) {
          QueryGrpc.getProposalsByGroupPolicyMethod = getProposalsByGroupPolicyMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest, cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ProposalsByGroupPolicy"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("ProposalsByGroupPolicy"))
              .build();
        }
      }
    }
    return getProposalsByGroupPolicyMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse> getVoteByProposalVoterMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "VoteByProposalVoter",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse> getVoteByProposalVoterMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest, cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse> getVoteByProposalVoterMethod;
    if ((getVoteByProposalVoterMethod = QueryGrpc.getVoteByProposalVoterMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getVoteByProposalVoterMethod = QueryGrpc.getVoteByProposalVoterMethod) == null) {
          QueryGrpc.getVoteByProposalVoterMethod = getVoteByProposalVoterMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest, cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "VoteByProposalVoter"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("VoteByProposalVoter"))
              .build();
        }
      }
    }
    return getVoteByProposalVoterMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse> getVotesByProposalMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "VotesByProposal",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse> getVotesByProposalMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest, cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse> getVotesByProposalMethod;
    if ((getVotesByProposalMethod = QueryGrpc.getVotesByProposalMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getVotesByProposalMethod = QueryGrpc.getVotesByProposalMethod) == null) {
          QueryGrpc.getVotesByProposalMethod = getVotesByProposalMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest, cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "VotesByProposal"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("VotesByProposal"))
              .build();
        }
      }
    }
    return getVotesByProposalMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse> getVotesByVoterMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "VotesByVoter",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse> getVotesByVoterMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest, cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse> getVotesByVoterMethod;
    if ((getVotesByVoterMethod = QueryGrpc.getVotesByVoterMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getVotesByVoterMethod = QueryGrpc.getVotesByVoterMethod) == null) {
          QueryGrpc.getVotesByVoterMethod = getVotesByVoterMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest, cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "VotesByVoter"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("VotesByVoter"))
              .build();
        }
      }
    }
    return getVotesByVoterMethod;
  }

  private static volatile io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse> getGroupsByMemberMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GroupsByMember",
      requestType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest.class,
      responseType = cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest,
      cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse> getGroupsByMemberMethod() {
    io.grpc.MethodDescriptor<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse> getGroupsByMemberMethod;
    if ((getGroupsByMemberMethod = QueryGrpc.getGroupsByMemberMethod) == null) {
      synchronized (QueryGrpc.class) {
        if ((getGroupsByMemberMethod = QueryGrpc.getGroupsByMemberMethod) == null) {
          QueryGrpc.getGroupsByMemberMethod = getGroupsByMemberMethod =
              io.grpc.MethodDescriptor.<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest, cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GroupsByMember"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryMethodDescriptorSupplier("GroupsByMember"))
              .build();
        }
      }
    }
    return getGroupsByMemberMethod;
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
   * Query is the cosmos.group.v1beta1 Query service.
   * </pre>
   */
  public static abstract class QueryImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * GroupInfo queries group info based on group id.
     * </pre>
     */
    public void groupInfo(cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGroupInfoMethod(), responseObserver);
    }

    /**
     * <pre>
     * GroupPolicyInfo queries group policy info based on account address of group policy.
     * </pre>
     */
    public void groupPolicyInfo(cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGroupPolicyInfoMethod(), responseObserver);
    }

    /**
     * <pre>
     * GroupMembers queries members of a group
     * </pre>
     */
    public void groupMembers(cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGroupMembersMethod(), responseObserver);
    }

    /**
     * <pre>
     * GroupsByAdmin queries groups by admin address.
     * </pre>
     */
    public void groupsByAdmin(cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGroupsByAdminMethod(), responseObserver);
    }

    /**
     * <pre>
     * GroupPoliciesByGroup queries group policies by group id.
     * </pre>
     */
    public void groupPoliciesByGroup(cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGroupPoliciesByGroupMethod(), responseObserver);
    }

    /**
     * <pre>
     * GroupsByAdmin queries group policies by admin address.
     * </pre>
     */
    public void groupPoliciesByAdmin(cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGroupPoliciesByAdminMethod(), responseObserver);
    }

    /**
     * <pre>
     * Proposal queries a proposal based on proposal id.
     * </pre>
     */
    public void proposal(cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getProposalMethod(), responseObserver);
    }

    /**
     * <pre>
     * ProposalsByGroupPolicy queries proposals based on account address of group policy.
     * </pre>
     */
    public void proposalsByGroupPolicy(cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getProposalsByGroupPolicyMethod(), responseObserver);
    }

    /**
     * <pre>
     * VoteByProposalVoter queries a vote by proposal id and voter.
     * </pre>
     */
    public void voteByProposalVoter(cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getVoteByProposalVoterMethod(), responseObserver);
    }

    /**
     * <pre>
     * VotesByProposal queries a vote by proposal.
     * </pre>
     */
    public void votesByProposal(cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getVotesByProposalMethod(), responseObserver);
    }

    /**
     * <pre>
     * VotesByVoter queries a vote by voter.
     * </pre>
     */
    public void votesByVoter(cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getVotesByVoterMethod(), responseObserver);
    }

    /**
     * <pre>
     * GroupsByMember queries groups by member address.
     * </pre>
     */
    public void groupsByMember(cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGroupsByMemberMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getGroupInfoMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse>(
                  this, METHODID_GROUP_INFO)))
          .addMethod(
            getGroupPolicyInfoMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse>(
                  this, METHODID_GROUP_POLICY_INFO)))
          .addMethod(
            getGroupMembersMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse>(
                  this, METHODID_GROUP_MEMBERS)))
          .addMethod(
            getGroupsByAdminMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse>(
                  this, METHODID_GROUPS_BY_ADMIN)))
          .addMethod(
            getGroupPoliciesByGroupMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse>(
                  this, METHODID_GROUP_POLICIES_BY_GROUP)))
          .addMethod(
            getGroupPoliciesByAdminMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse>(
                  this, METHODID_GROUP_POLICIES_BY_ADMIN)))
          .addMethod(
            getProposalMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse>(
                  this, METHODID_PROPOSAL)))
          .addMethod(
            getProposalsByGroupPolicyMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse>(
                  this, METHODID_PROPOSALS_BY_GROUP_POLICY)))
          .addMethod(
            getVoteByProposalVoterMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse>(
                  this, METHODID_VOTE_BY_PROPOSAL_VOTER)))
          .addMethod(
            getVotesByProposalMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse>(
                  this, METHODID_VOTES_BY_PROPOSAL)))
          .addMethod(
            getVotesByVoterMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse>(
                  this, METHODID_VOTES_BY_VOTER)))
          .addMethod(
            getGroupsByMemberMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest,
                cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse>(
                  this, METHODID_GROUPS_BY_MEMBER)))
          .build();
    }
  }

  /**
   * <pre>
   * Query is the cosmos.group.v1beta1 Query service.
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
     * GroupInfo queries group info based on group id.
     * </pre>
     */
    public void groupInfo(cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGroupInfoMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GroupPolicyInfo queries group policy info based on account address of group policy.
     * </pre>
     */
    public void groupPolicyInfo(cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGroupPolicyInfoMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GroupMembers queries members of a group
     * </pre>
     */
    public void groupMembers(cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGroupMembersMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GroupsByAdmin queries groups by admin address.
     * </pre>
     */
    public void groupsByAdmin(cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGroupsByAdminMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GroupPoliciesByGroup queries group policies by group id.
     * </pre>
     */
    public void groupPoliciesByGroup(cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGroupPoliciesByGroupMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GroupsByAdmin queries group policies by admin address.
     * </pre>
     */
    public void groupPoliciesByAdmin(cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGroupPoliciesByAdminMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Proposal queries a proposal based on proposal id.
     * </pre>
     */
    public void proposal(cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getProposalMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * ProposalsByGroupPolicy queries proposals based on account address of group policy.
     * </pre>
     */
    public void proposalsByGroupPolicy(cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getProposalsByGroupPolicyMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * VoteByProposalVoter queries a vote by proposal id and voter.
     * </pre>
     */
    public void voteByProposalVoter(cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getVoteByProposalVoterMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * VotesByProposal queries a vote by proposal.
     * </pre>
     */
    public void votesByProposal(cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getVotesByProposalMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * VotesByVoter queries a vote by voter.
     * </pre>
     */
    public void votesByVoter(cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getVotesByVoterMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GroupsByMember queries groups by member address.
     * </pre>
     */
    public void groupsByMember(cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest request,
        io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGroupsByMemberMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * <pre>
   * Query is the cosmos.group.v1beta1 Query service.
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
     * GroupInfo queries group info based on group id.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse groupInfo(cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGroupInfoMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GroupPolicyInfo queries group policy info based on account address of group policy.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse groupPolicyInfo(cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGroupPolicyInfoMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GroupMembers queries members of a group
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse groupMembers(cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGroupMembersMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GroupsByAdmin queries groups by admin address.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse groupsByAdmin(cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGroupsByAdminMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GroupPoliciesByGroup queries group policies by group id.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse groupPoliciesByGroup(cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGroupPoliciesByGroupMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GroupsByAdmin queries group policies by admin address.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse groupPoliciesByAdmin(cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGroupPoliciesByAdminMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Proposal queries a proposal based on proposal id.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse proposal(cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getProposalMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * ProposalsByGroupPolicy queries proposals based on account address of group policy.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse proposalsByGroupPolicy(cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getProposalsByGroupPolicyMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * VoteByProposalVoter queries a vote by proposal id and voter.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse voteByProposalVoter(cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getVoteByProposalVoterMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * VotesByProposal queries a vote by proposal.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse votesByProposal(cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getVotesByProposalMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * VotesByVoter queries a vote by voter.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse votesByVoter(cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getVotesByVoterMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GroupsByMember queries groups by member address.
     * </pre>
     */
    public cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse groupsByMember(cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGroupsByMemberMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * Query is the cosmos.group.v1beta1 Query service.
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
     * GroupInfo queries group info based on group id.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse> groupInfo(
        cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGroupInfoMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GroupPolicyInfo queries group policy info based on account address of group policy.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse> groupPolicyInfo(
        cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGroupPolicyInfoMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GroupMembers queries members of a group
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse> groupMembers(
        cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGroupMembersMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GroupsByAdmin queries groups by admin address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse> groupsByAdmin(
        cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGroupsByAdminMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GroupPoliciesByGroup queries group policies by group id.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse> groupPoliciesByGroup(
        cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGroupPoliciesByGroupMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GroupsByAdmin queries group policies by admin address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse> groupPoliciesByAdmin(
        cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGroupPoliciesByAdminMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Proposal queries a proposal based on proposal id.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse> proposal(
        cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getProposalMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * ProposalsByGroupPolicy queries proposals based on account address of group policy.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse> proposalsByGroupPolicy(
        cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getProposalsByGroupPolicyMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * VoteByProposalVoter queries a vote by proposal id and voter.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse> voteByProposalVoter(
        cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getVoteByProposalVoterMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * VotesByProposal queries a vote by proposal.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse> votesByProposal(
        cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getVotesByProposalMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * VotesByVoter queries a vote by voter.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse> votesByVoter(
        cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getVotesByVoterMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GroupsByMember queries groups by member address.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse> groupsByMember(
        cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGroupsByMemberMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_GROUP_INFO = 0;
  private static final int METHODID_GROUP_POLICY_INFO = 1;
  private static final int METHODID_GROUP_MEMBERS = 2;
  private static final int METHODID_GROUPS_BY_ADMIN = 3;
  private static final int METHODID_GROUP_POLICIES_BY_GROUP = 4;
  private static final int METHODID_GROUP_POLICIES_BY_ADMIN = 5;
  private static final int METHODID_PROPOSAL = 6;
  private static final int METHODID_PROPOSALS_BY_GROUP_POLICY = 7;
  private static final int METHODID_VOTE_BY_PROPOSAL_VOTER = 8;
  private static final int METHODID_VOTES_BY_PROPOSAL = 9;
  private static final int METHODID_VOTES_BY_VOTER = 10;
  private static final int METHODID_GROUPS_BY_MEMBER = 11;

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
        case METHODID_GROUP_INFO:
          serviceImpl.groupInfo((cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupInfoResponse>) responseObserver);
          break;
        case METHODID_GROUP_POLICY_INFO:
          serviceImpl.groupPolicyInfo((cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPolicyInfoResponse>) responseObserver);
          break;
        case METHODID_GROUP_MEMBERS:
          serviceImpl.groupMembers((cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupMembersResponse>) responseObserver);
          break;
        case METHODID_GROUPS_BY_ADMIN:
          serviceImpl.groupsByAdmin((cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByAdminResponse>) responseObserver);
          break;
        case METHODID_GROUP_POLICIES_BY_GROUP:
          serviceImpl.groupPoliciesByGroup((cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByGroupResponse>) responseObserver);
          break;
        case METHODID_GROUP_POLICIES_BY_ADMIN:
          serviceImpl.groupPoliciesByAdmin((cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupPoliciesByAdminResponse>) responseObserver);
          break;
        case METHODID_PROPOSAL:
          serviceImpl.proposal((cosmos.group.v1beta1.QueryOuterClass.QueryProposalRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryProposalResponse>) responseObserver);
          break;
        case METHODID_PROPOSALS_BY_GROUP_POLICY:
          serviceImpl.proposalsByGroupPolicy((cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryProposalsByGroupPolicyResponse>) responseObserver);
          break;
        case METHODID_VOTE_BY_PROPOSAL_VOTER:
          serviceImpl.voteByProposalVoter((cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryVoteByProposalVoterResponse>) responseObserver);
          break;
        case METHODID_VOTES_BY_PROPOSAL:
          serviceImpl.votesByProposal((cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByProposalResponse>) responseObserver);
          break;
        case METHODID_VOTES_BY_VOTER:
          serviceImpl.votesByVoter((cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryVotesByVoterResponse>) responseObserver);
          break;
        case METHODID_GROUPS_BY_MEMBER:
          serviceImpl.groupsByMember((cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberRequest) request,
              (io.grpc.stub.StreamObserver<cosmos.group.v1beta1.QueryOuterClass.QueryGroupsByMemberResponse>) responseObserver);
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
      return cosmos.group.v1beta1.QueryOuterClass.getDescriptor();
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
              .addMethod(getGroupInfoMethod())
              .addMethod(getGroupPolicyInfoMethod())
              .addMethod(getGroupMembersMethod())
              .addMethod(getGroupsByAdminMethod())
              .addMethod(getGroupPoliciesByGroupMethod())
              .addMethod(getGroupPoliciesByAdminMethod())
              .addMethod(getProposalMethod())
              .addMethod(getProposalsByGroupPolicyMethod())
              .addMethod(getVoteByProposalVoterMethod())
              .addMethod(getVotesByProposalMethod())
              .addMethod(getVotesByVoterMethod())
              .addMethod(getGroupsByMemberMethod())
              .build();
        }
      }
    }
    return result;
  }
}
