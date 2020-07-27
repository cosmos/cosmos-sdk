package enforceproto

import (
	"reflect"
	"testing"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestCheckExtraneousFieldsRepeated(t *testing.T) {
	tests := []struct {
		name    string
		in      proto.Message
		recv    proto.Message
		wantErr error
	}{
		{
			name: "TestVersionFD1 vs TestVersion3 -- full equivalence",
			in: &testdata.TestVersionFD1{
				H: []*testdata.TestVersion1{
					{
						H: []*testdata.TestVersion1{
							{
								Sum: &testdata.TestVersion1_F{
									F: &testdata.TestVersion1{
										A: &testdata.TestVersion1{
											B: &testdata.TestVersion1{
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
						},
					},
				},
			},
			recv:    new(testdata.TestVersion3),
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
			gotErr := CheckMismatchedProtoFields(protoBlob, tt.recv)
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Fatalf("Error mismatch\nGot:\n%v\n\nWant:\n%v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestCheckExtraneousFieldsNested(t *testing.T) {
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
			wantErr: &errExtraneousField{
				Type:     "*testdata.TestVersionFD1",
				TagNum:   12,
				WireType: 2,
			},
		},
		{
			name: "Nested2B vs Nested3B",
			in: &testdata.Nested2B{
				Id:  5,
				Fee: 88.91,
				Nested: &testdata.Nested3B{
					Id:   15,
					Age:  47,
					Name: "Crowley",
					B4: []*testdata.Nested4B{
						{
							Id: 34, Age: 55, Name: "Trivia",
						},
						{
							Id: 55, Age: 78, Name: "Incredible",
						},
					},
				},
			},
			recv: new(testdata.TestVersion3),
			wantErr: &errMismatchedWireType{
				Type:         "*testdata.TestVersion3",
				TagNum:       2,
				GotWireType:  1,
				WantWireType: 2,
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
			wantErr: &errExtraneousField{
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
			wantErr: &errExtraneousField{
				Type:     "*testdata.TestVersion3LoneNesting",
				TagNum:   6,
				WireType: 0,
			},
		},
		{
			name: "unknown field types.Any in G",
			in: &testdata.TestVersion3{
				G: &types.Any{
					TypeUrl: "/testdata.TestVersion1",
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
			wantErr: &errExtraneousField{
				Type:   "*testdata.TestVersion1",
				TagNum: 25,
			},
		},
		{
			name: "mismatched types.Any in G",
			in: &testdata.TestVersion1{
				G: &types.Any{
					TypeUrl: "/testdata.TestVersion4LoneNesting",
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
				TagNum:       1,
				GotWireType:  2,
				WantWireType: 0,
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
			recv: new(testdata.TestVersion4LoneNesting),
			wantErr: &errMismatchedWireType{
				Type:         "*testdata.TestVersion4LoneNesting_Inner1_InnerInner",
				TagNum:       1,
				GotWireType:  2,
				WantWireType: 0,
			},
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
			recv: new(testdata.TestVersion4LoneNesting),
			wantErr: &errMismatchedWireType{
				Type:         "*testdata.TestVersion4LoneNesting_Inner2_InnerInner",
				TagNum:       2,
				GotWireType:  2,
				WantWireType: 0,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			protoBlob, err := proto.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			gotErr := CheckMismatchedProtoFields(protoBlob, tt.recv)
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Fatalf("Error mismatch\nGot:\n%s\n\nWant:\n%s", gotErr, tt.wantErr)
			}
		})
	}
}

func mustMarshal(msg proto.Message) []byte {
	blob, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return blob
}

func TestCheckExtraneousFieldsFlat(t *testing.T) {
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
			wantErr: &errExtraneousField{
				Type:   "*testdata.Customer1",
				TagNum: 8, WireType: 2,
			},
		},
		{
			name: "Any in a field, the extra field will be serialized so should fail",
			in: &testdata.Customer2{
				Miscellaneous: &types.Any{},
			},
			wantErr: &errExtraneousField{
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
			wantErr: &errExtraneousField{
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
			name: "Using a field that's in the reserved range, should pass",
			in: &testdata.Customer2{
				Id:       289,
				Reserved: 99,
			},
			wantErr: nil,
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
			gotErr := CheckMismatchedProtoFields(blob, c1)
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Fatalf("Error mismatch\nGot:\n%s\n\nWant:\n%s", gotErr, tt.wantErr)
			}
		})
	}
}
