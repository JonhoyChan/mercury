package api

import (
	"outgoing/x/ecode"
)

func NewProto(operation int32) *Proto {
	return &Proto{
		Version:   ProtocolVersion,
		Operation: operation,
		Body:      []byte{},
	}
}

func NewResponse(err error, mid string, timestamp int64) *Response {
	code := ecode.Cause(err)
	return &Response{
		MID:       mid,
		Code:      code.Code(),
		Message:   code.Message(),
		Timestamp: timestamp,
	}
}
