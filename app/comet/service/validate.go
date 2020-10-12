package service

import (
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"mercury/x"
	"mercury/x/types"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("is-receiver", ValidateReceiver); err != nil {
		panic(x.Sprintf("unable to register %s validation: %s", "is-receiver", err.Error()))
	}
	if err := validate.RegisterValidation("is-body", ValidateBody); err != nil {
		panic(x.Sprintf("unable to register %s validation: %s", "is-body", err.Error()))
	}
}

func ValidateReceiver(fl validator.FieldLevel) bool {
	var id types.ID
	switch types.MessageType(fl.Parent().Elem().FieldByName("MessageType").Uint()) {
	case types.MessageTypeSingle:
		id = types.ParseUID(fl.Field().String())
	case types.MessageTypeGroup:
		id = types.ParseGID(fl.Field().String())
	default:
		return false
	}
	return !id.IsZero()
}

func ValidateBody(fl validator.FieldLevel) bool {
	var v interface{}
	switch types.ContentType(fl.Parent().Elem().FieldByName("ContentType").Uint()) {
	case types.ContentTypeText:
		v = &(types.TextMessage{})
	case types.ContentTypeImage:
		v = &(types.ImageMessage{})
	case types.ContentTypeLocation:
		v = &(types.LocationMessage{})
	case types.ContentTypeAudio:
		v = &(types.AudioMessage{})
	case types.ContentTypeVideo:
		v = &(types.VideoMessage{})
	case types.ContentTypeFile:
		v = &(types.FileMessage{})
	}
	err := jsoniter.Unmarshal(fl.Field().Bytes(), v)
	if err != nil {
		return false
	}
	err = validate.Struct(v)
	return err == nil
}
