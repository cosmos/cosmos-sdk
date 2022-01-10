package group_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

var (
	admin   = sdk.AccAddress("admin")
	member1 = sdk.AccAddress("member1")
	member2 = sdk.AccAddress("member2")
)

func TestMsgCreateGroup(t *testing.T) {
	testCases := []struct {
		name   string
		msg    *group.MsgCreateGroup
		expErr bool
		errMsg string
	}{
		{
			"invalid admin",
			&group.MsgCreateGroup{
				Admin: "invalid admin",
			},
			true,
			"admin: decoding bech32 failed",
		},
		{
			"invalid member address",
			&group.MsgCreateGroup{
				Admin: admin.String(),
				Members: []group.Member{
					group.Member{
						Address: "invalid address",
					},
				},
			},
			true,
			"members: address: decoding bech32 failed",
		},
		{
			"negitive member weight",
			&group.MsgCreateGroup{
				Admin: admin.String(),
				Members: []group.Member{
					group.Member{
						Address: member1.String(),
						Weight:  "-1",
					},
				},
			},
			true,
			"expected a positive decimal",
		},
		{
			"duplicate member",
			&group.MsgCreateGroup{
				Admin: admin.String(),
				Members: []group.Member{
					group.Member{
						Address:  member1.String(),
						Weight:   "1",
						Metadata: []byte("metadata"),
					},
					group.Member{
						Address:  member1.String(),
						Weight:   "1",
						Metadata: []byte("metadata"),
					},
				},
			},
			true,
			"duplicate value",
		},
		{
			"valid test case",
			&group.MsgCreateGroup{
				Admin: admin.String(),
				Members: []group.Member{
					group.Member{
						Address:  member1.String(),
						Weight:   "1",
						Metadata: []byte("metadata"),
					},
				},
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.msg.Type(), group.TypeMsgCreateGroup)
			}
		})
	}
}

func TestMsgUpdateGroupAdmin(t *testing.T) {
	testCases := []struct {
		name   string
		msg    *group.MsgUpdateGroupAdmin
		expErr bool
		errMsg string
	}{
		{
			"empty group id",
			&group.MsgUpdateGroupAdmin{
				Admin:    admin.String(),
				NewAdmin: member1.String(),
			},
			true,
			"group-id: value is empty",
		},
		{
			"admin: invalid bech32 address",
			&group.MsgUpdateGroupAdmin{
				GroupId: 1,
				Admin:   "admin",
			},
			true,
			"admin: decoding bech32 failed",
		},
		{
			"new admin: invalid bech32 address",
			&group.MsgUpdateGroupAdmin{
				GroupId:  1,
				Admin:    admin.String(),
				NewAdmin: "new-admin",
			},
			true,
			"new admin: decoding bech32 failed",
		},
		{
			"admin & new admin is same",
			&group.MsgUpdateGroupAdmin{
				GroupId:  1,
				Admin:    admin.String(),
				NewAdmin: admin.String(),
			},
			true,
			"new and old admin are the same",
		},
		{
			"valid case",
			&group.MsgUpdateGroupAdmin{
				GroupId:  1,
				Admin:    admin.String(),
				NewAdmin: member1.String(),
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.msg.Type(), group.TypeMsgUpdateGroupAdmin)
			}
		})
	}
}

func TestMsgUpdateGroupMetadata(t *testing.T) {
	testCases := []struct {
		name   string
		msg    *group.MsgUpdateGroupMetadata
		expErr bool
		errMsg string
	}{
		{
			"empty group id",
			&group.MsgUpdateGroupMetadata{
				Admin: admin.String(),
			},
			true,
			"group-id: value is empty",
		},
		{
			"admin: invalid bech32 address",
			&group.MsgUpdateGroupMetadata{
				GroupId: 1,
				Admin:   "admin",
			},
			true,
			"admin: decoding bech32 failed",
		},
		{
			"valid test",
			&group.MsgUpdateGroupMetadata{
				GroupId:  1,
				Admin:    admin.String(),
				Metadata: []byte("metadata"),
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.msg.Type(), group.TypeMsgUpdateGroupMetadata)
			}
		})
	}
}
