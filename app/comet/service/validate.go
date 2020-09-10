package service

import (
	"github.com/go-playground/validator/v10"
	"outgoing/x"
	"outgoing/x/types"
	"strings"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("is-id", ValidateID); err != nil {
		panic(x.Sprintf("unable to register %s validation: %s", "is-id", err.Error()))
	}
}

func ValidateID(fl validator.FieldLevel) bool {
	var id types.ID
	s := fl.Field().String()
	if strings.HasPrefix(s, types.PrefixUID) {
		id = types.ParseUID(s)
	} else if strings.HasPrefix(s, types.PrefixGID) {
		id = types.ParseGID(s)
	} else {
		return false
	}
	return !id.IsZero()
}
