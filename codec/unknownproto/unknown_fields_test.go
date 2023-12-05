package unknownproto

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestRejectUnknownFieldsRepeated(t *testing.T) {
	tests := []struct {
		name                     string
		in                       proto.Message
		recv                     proto.Message
		wantErr                  error
		allowUnknownNonCriticals bool
		hasUnknownNonCriticals   bool
	}{
		{
			name: "Unknown field in midst of repeated values",
			in: &testdata.TestVersion2{
				C: []*testdata.TestVersion2{
					{
						C: []*testdata.TestVersion2{
							{
								Sum: &testdata.TestVersion2_F{
									F: &testdata.TestVersion2{
										A: &testdata.TestVersion2{
											B: &testdata.TestVersion2{
												H: []*testdata.TestVersion1{
													{
														X: 0x01,
													},
												},
											},
										},
									},
								},
							},
							{
								Sum: &testdata.TestVersion2_F{
									F: &testdata.TestVersion2{
										A: &testdata.TestVersion2{
											B: &testdata.TestVersion2{
												H: []*testdata.TestVersion1{
													{
														X: 0x02,
													},
												},
											},
										},
									},
								},
							},
							{
								Sum: &testdata.TestVersion2_F{
									F: &testdata.TestVersion2{
										NewField: 411,
									},
								},
							},
						},
					},
				},
			},
			recv: new(testdata.TestVersion1),
			wantErr: &errUnknownField{
				Type:     "*testdata.TestVersion1",
				TagNum:   25,
				WireType: 0,
			},
		},
		{
			name:                     "Unknown field in midst of repeated values, allowUnknownNonCriticals set",
			allowUnknownNonCriticals: true,
			in: &testdata.TestVersion2{
				C: []*testdata.TestVersion2{
					{
						C: []*testdata.TestVersion2{
							{
								Sum: &testdata.TestVersion2_F{
									F: &testdata.TestVersion2{
										A: &testdata.TestVersion2{
											B: &testdata.TestVersion2{
												H: []*testdata.TestVersion1{
													{
														X: 0x01,
													},
												},
											},
										},
									},
								},
							},
							{
								Sum: &testdata.TestVersion2_F{
									F: &testdata.TestVersion2{
										A: &testdata.TestVersion2{
											B: &testdata.TestVersion2{
												H: []*testdata.TestVersion1{
													{
														X: 0x02,
													},
												},
											},
										},
									},
								},
							},
							{
								Sum: &testdata.TestVersion2_F{
									F: &testdata.TestVersion2{
										NewField: 411,
									},
								},
							},
						},
					},
				},
			},
			recv: new(testdata.TestVersion1),
			wantErr: &errUnknownField{
				Type:     "*testdata.TestVersion1",
				TagNum:   25,
				WireType: 0,
			},
		},
		{
			name: "Unknown field in midst of repeated values, non-critical field to be rejected",
			in: &testdata.TestVersion3{
				C: []*testdata.TestVersion3{
					{
						C: []*testdata.TestVersion3{
							{
								Sum: &testdata.TestVersion3_F{
									F: &testdata.TestVersion3{
										A: &testdata.TestVersion3{
											B: &testdata.TestVersion3{
												X: 0x01,
											},
										},
									},
								},
							},
							{
								Sum: &testdata.TestVersion3_F{
									F: &testdata.TestVersion3{
										A: &testdata.TestVersion3{
											B: &testdata.TestVersion3{
												X: 0x02,
											},
										},
									},
								},
							},
							{
								Sum: &testdata.TestVersion3_F{
									F: &testdata.TestVersion3{
										NonCriticalField: "non-critical",
									},
								},
							},
						},
					},
				},
			},
			recv: new(testdata.TestVersion1),
			wantErr: &errUnknownField{
				Type:     "*testdata.TestVersion1",
				TagNum:   1031,
				WireType: 2,
			},
			hasUnknownNonCriticals: true,
		},
		{
			name:                     "Unknown field in midst of repeated values, non-critical field ignored",
			allowUnknownNonCriticals: true,
			in: &testdata.TestVersion3{
				C: []*testdata.TestVersion3{
					{
						C: []*testdata.TestVersion3{
							{
								Sum: &testdata.TestVersion3_F{
									F: &testdata.TestVersion3{
										A: &testdata.TestVersion3{
											B: &testdata.TestVersion3{
												X: 0x01,
											},
										},
									},
								},
							},
							{
								Sum: &testdata.TestVersion3_F{
									F: &testdata.TestVersion3{
										A: &testdata.TestVersion3{
											B: &testdata.TestVersion3{
												X: 0x02,
											},
										},
									},
								},
							},
							{
								Sum: &testdata.TestVersion3_F{
									F: &testdata.TestVersion3{
										NonCriticalField: "non-critical",
									},
								},
							},
						},
					},
				},
			},
			recv:                   new(testdata.TestVersion1),
			wantErr:                nil,
			hasUnknownNonCriticals: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			protoBlob, err := proto.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			hasUnknownNonCriticals, gotErr := RejectUnknownFields(protoBlob, tt.recv, tt.allowUnknownNonCriticals, DefaultAnyResolver{})
			require.Equal(t, tt.wantErr, gotErr)
			require.Equal(t, tt.hasUnknownNonCriticals, hasUnknownNonCriticals)
		})
	}
}

func TestRejectUnknownFields_allowUnknownNonCriticals(t *testing.T) {
	tests := []struct {
		name                     string
		in                       proto.Message
		allowUnknownNonCriticals bool
		wantErr                  error
	}{
		{
			name: "Field that's in the reserved range, should fail by default",
			in: &testdata.Customer2{
				Id:       289,
				Reserved: 99,
			},
			wantErr: &errUnknownField{
				Type:     "*testdata.Customer1",
				TagNum:   1047,
				WireType: 0,
			},
		},
		{
			name:                     "Field that's in the reserved range, toggle allowUnknownNonCriticals",
			allowUnknownNonCriticals: true,
			in: &testdata.Customer2{
				Id:       289,
				Reserved: 99,
			},
			wantErr: nil,
		},
		{
			name:                     "Unknown fields that are critical, but with allowUnknownNonCriticals set",
			allowUnknownNonCriticals: true,
			in: &testdata.Customer2{
				Id:   289,
				City: testdata.Customer2_PaloAlto,
			},
			wantErr: &errUnknownField{
				Type:     "*testdata.Customer1",
				TagNum:   6,
				WireType: 0,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			blob, err := proto.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			c1 := new(testdata.Customer1)
			_, gotErr := RejectUnknownFields(blob, c1, tt.allowUnknownNonCriticals, DefaultAnyResolver{})
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Fatalf("Error mismatch\nGot:\n%s\n\nWant:\n%s", gotErr, tt.wantErr)
			}
		})
	}
}

func TestRejectUnknownFieldsNested(t *testing.T) {
	tests := []struct {
		name    string
		in      proto.Message
		recv    proto.Message
		wantErr error
	}{
		{
			name: "TestVersion3 from TestVersionFD1",
			in: &testdata.TestVersion2{
				X: 5,
				Sum: &testdata.TestVersion2_E{
					E: 100,
				},
				H: []*testdata.TestVersion1{
					{X: 999},
					{X: -55},
					{
						X: 102,
						Sum: &testdata.TestVersion1_F{
							F: &testdata.TestVersion1{
								X: 4,
							},
						},
					},
				},
				Customer1: &testdata.Customer1{
					Id:              45,
					Name:            "customer1",
					SubscriptionFee: 99,
				},
			},
			recv: new(testdata.TestVersionFD1),
			wantErr: &errUnknownField{
				Type:     "*testdata.TestVersionFD1",
				TagNum:   12,
				WireType: 2,
			},
		},
		{
			name: "Alternating oneofs",
			in: &testdata.TestVersion3{
				Sum: &testdata.TestVersion3_E{
					E: 99,
				},
			},
			recv:    new(testdata.TestVersion3LoneOneOfValue),
			wantErr: nil,
		},
		{
			name: "Alternating oneofs mismatched field",
			in: &testdata.TestVersion3{
				Sum: &testdata.TestVersion3_F{
					F: &testdata.TestVersion3{
						X: 99,
					},
				},
			},
			recv: new(testdata.TestVersion3LoneOneOfValue),
			wantErr: &errUnknownField{
				Type:     "*testdata.TestVersion3LoneOneOfValue",
				TagNum:   7,
				WireType: 2,
			},
		},
		{
			name: "Discrepancy in a deeply nested one of field",
			in: &testdata.TestVersion3{
				Sum: &testdata.TestVersion3_F{
					F: &testdata.TestVersion3{
						Sum: &testdata.TestVersion3_F{
							F: &testdata.TestVersion3{
								X: 19,
								Sum: &testdata.TestVersion3_E{
									E: 99,
								},
							},
						},
					},
				},
			},
			recv: new(testdata.TestVersion3LoneNesting),
			wantErr: &errUnknownField{
				Type:     "*testdata.TestVersion3LoneNesting",
				TagNum:   6,
				WireType: 0,
			},
		},
		{
			name: "unknown field types.Any in G",
			in: &testdata.TestVersion3{
				G: &types.Any{
					TypeUrl: "/testpb.TestVersion1",
					Value: mustMarshal(&testdata.TestVersion2{
						Sum: &testdata.TestVersion2_F{
							F: &testdata.TestVersion2{
								NewField: 999,
							},
						},
					}),
				},
			},
			recv: new(testdata.TestVersion3),
			wantErr: &errUnknownField{
				Type:   "*testdata.TestVersion1",
				TagNum: 25,
			},
		},
		{
			name: "types.Any with extra fields",
			in: &testdata.TestVersionFD1WithExtraAny{
				G: &testdata.AnyWithExtra{
					Any: &types.Any{
						TypeUrl: "/testpb.TestVersion1",
						Value: mustMarshal(&testdata.TestVersion2{
							Sum: &testdata.TestVersion2_F{
								F: &testdata.TestVersion2{
									NewField: 999,
								},
							},
						}),
					},
					B: 3,
					C: 2,
				},
			},
			recv: new(testdata.TestVersion3),
			wantErr: &errUnknownField{
				Type:     "*types.Any",
				TagNum:   3,
				WireType: 0,
			},
		},
		{
			name: "mismatched types.Any in G",
			in: &testdata.TestVersion1{
				G: &types.Any{
					TypeUrl: "/testpb.TestVersion4LoneNesting",
					Value: mustMarshal(&testdata.TestVersion3LoneNesting_Inner1{
						Inner: &testdata.TestVersion3LoneNesting_Inner1_InnerInner{
							Id:   "ID",
							City: "Gotham",
						},
					}),
				},
			},
			recv: new(testdata.TestVersion1),
			wantErr: &errMismatchedWireType{
				Type:         "*testdata.TestVersion3",
				TagNum:       8,
				GotWireType:  7,
				WantWireType: 2,
			},
		},
		{
			name: "From nested proto message, message index 0",
			in: &testdata.TestVersion3LoneNesting{
				Inner1: &testdata.TestVersion3LoneNesting_Inner1{
					Id:   10,
					Name: "foo",
					Inner: &testdata.TestVersion3LoneNesting_Inner1_InnerInner{
						Id:   "ID",
						City: "Palo Alto",
					},
				},
			},
			recv:    new(testdata.TestVersion4LoneNesting),
			wantErr: nil,
		},
		{
			name: "From nested proto message, message index 1",
			in: &testdata.TestVersion3LoneNesting{
				Inner2: &testdata.TestVersion3LoneNesting_Inner2{
					Id:      "ID",
					Country: "Maldives",
					Inner: &testdata.TestVersion3LoneNesting_Inner2_InnerInner{
						Id:   "ID",
						City: "Unknown",
					},
				},
			},
			recv:    new(testdata.TestVersion4LoneNesting),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			protoBlob, err := proto.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			gotErr := RejectUnknownFieldsStrict(protoBlob, tt.recv, DefaultAnyResolver{})
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Fatalf("Error mismatch\nGot:\n%s\n\nWant:\n%s", gotErr, tt.wantErr)
			}
		})
	}
}

func TestRejectUnknownFieldsFlat(t *testing.T) {
	tests := []struct {
		name    string
		in      proto.Message
		wantErr error
	}{
		{
			name: "Oneof with same field number, shouldn't complain",
			in: &testdata.Customer3{
				Id:   68,
				Name: "ACME3",
				Payment: &testdata.Customer3_CreditCardNo{
					CreditCardNo: "123-XXXX-XXX881",
				},
			},
			wantErr: nil,
		},
		{
			name: "Oneof with different field number, should fail",
			in: &testdata.Customer3{
				Id:   68,
				Name: "ACME3",
				Payment: &testdata.Customer3_ChequeNo{
					ChequeNo: "123XXXXXXX881",
				},
			},
			wantErr: &errUnknownField{
				Type:   "*testdata.Customer1",
				TagNum: 8, WireType: 2,
			},
		},
		{
			name: "Any in a field, the extra field will be serialized so should fail",
			in: &testdata.Customer2{
				Miscellaneous: &types.Any{},
			},
			wantErr: &errUnknownField{
				Type:     "*testdata.Customer1",
				TagNum:   10,
				WireType: 2,
			},
		},
		{
			name: "With a nested struct as a field",
			in: &testdata.Customer3{
				Id: 289,
				Original: &testdata.Customer1{
					Id: 991,
				},
			},
			wantErr: &errUnknownField{
				Type:     "*testdata.Customer1",
				TagNum:   9,
				WireType: 2,
			},
		},
		{
			name: "An extra field that's non-existent in Customer1",
			in: &testdata.Customer2{
				Id:       289,
				Name:     "Customer1",
				Industry: 5299,
				Fewer:    199.9,
			},
			wantErr: &errMismatchedWireType{
				Type:   "*testdata.Customer1",
				TagNum: 2, GotWireType: 0, WantWireType: 2,
			},
		},
		{
			name: "Using a field that's in the reserved range, should fail by default",
			in: &testdata.Customer2{
				Id:       289,
				Reserved: 99,
			},
			wantErr: &errUnknownField{
				Type:     "*testdata.Customer1",
				TagNum:   1047,
				WireType: 0,
			},
		},
		{
			name: "Only fields matching",
			in: &testdata.Customer2{
				Id:   289,
				Name: "Customer1",
			},
			wantErr: &errMismatchedWireType{
				Type:   "*testdata.Customer1",
				TagNum: 3, GotWireType: 2, WantWireType: 5,
			},
		},
		{
			name: "Extra field that's non-existent in Customer1, along with Reserved set",
			in: &testdata.Customer2{
				Id:       289,
				Name:     "Customer1",
				Industry: 5299,
				Fewer:    199.9,
				Reserved: 819,
			},
			wantErr: &errMismatchedWireType{
				Type:   "*testdata.Customer1",
				TagNum: 2, GotWireType: 0, WantWireType: 2,
			},
		},
		{
			name: "Using enumerated field",
			in: &testdata.Customer2{
				Id:       289,
				Name:     "Customer1",
				Industry: 5299,
				City:     testdata.Customer2_PaloAlto,
			},
			wantErr: &errMismatchedWireType{
				Type:        "*testdata.Customer1",
				TagNum:      2,
				GotWireType: 0, WantWireType: 2,
			},
		},
		{
			name: "multiple extraneous fields",
			in: &testdata.Customer2{
				Id:       289,
				Name:     "Customer1",
				Industry: 5299,
				City:     testdata.Customer2_PaloAlto,
				Fewer:    45,
			},
			wantErr: &errMismatchedWireType{
				TagNum: 2, GotWireType: 0, WantWireType: 2,
				Type: "*testdata.Customer1",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			blob, err := proto.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			c1 := new(testdata.Customer1)
			gotErr := RejectUnknownFieldsStrict(blob, c1, DefaultAnyResolver{})
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Fatalf("Error mismatch\nGot:\n%s\n\nWant:\n%s", gotErr, tt.wantErr)
			}
		})
	}
}

// Issue https://github.com/cosmos/cosmos-sdk/issues/7222, we need to ensure that repeated
// uint64 are recognized as packed.
func TestPackedEncoding(t *testing.T) {
	data := testdata.TestRepeatedUints{Nums: []uint64{12, 13}}

	marshaled, err := data.Marshal()
	require.NoError(t, err)

	unmarshalled := &testdata.TestRepeatedUints{}
	_, err = RejectUnknownFields(marshaled, unmarshalled, false, DefaultAnyResolver{})
	require.NoError(t, err)
}

func mustMarshal(msg proto.Message) []byte {
	blob, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return blob
}

// mockRegistry is an unfriendly registry which does not resolve any types.
// It is used to mock the case in which the interface registry does not know
// about a type contained in an Any.
type mockRegistry struct {
	protoregistry.MessageTypeResolver
}

func (mockRegistry) Resolve(typeUrl string) (proto.Message, error) {
	return nil, fmt.Errorf("always failing")
}

// TestUnknownFieldsAny tests that if the registry is provided with an auxiliary protov2 resolver
// for unknown types within any's then it will use them to resolve the type.
func TestUnknownFieldsAny(t *testing.T) {
	registry := mockRegistry{protoregistry.GlobalTypes}

	anyMsg, err := anypb.New(&testpb.Bird{
		Species: "cool-bird",
		Color:   100,
	})
	require.NoError(t, err)

	msgSent := &testdata.TestVersion3{
		G: &types.Any{
			TypeUrl: anyMsg.TypeUrl,
			Value:   anyMsg.Value,
		},
	}

	msgBytes := mustMarshal(msgSent)
	msgWant := new(testdata.TestVersion3)

	err = RejectUnknownFieldsStrict(msgBytes, msgWant, registry)
	require.NoError(t, err)
}
