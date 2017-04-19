package server

import (
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
)

var v = validator.New()

func validate(req interface{}) error {
	return errors.Wrap(v.Struct(req), "Validate")
}
