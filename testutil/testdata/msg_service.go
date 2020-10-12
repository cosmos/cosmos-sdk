package testdata

import (
	"context"
)

type MsgImpl struct{}

var _ MsgServer = MsgImpl

// CreateDog implements the MsgServer interface.
func (m MsgImpl) CreateDog(_ context.Context, msg *MsgCreateDog) (*MsgCreateDogResponse, error) {
	return &MsgCreateDogResponse{
		Name: msg.Dog.Name,
	}
}
