package service

import (
	jsoniter "github.com/json-iterator/go"
	"mercury/x/ecode"
)

type Response struct {
	// client-provided message id
	MID string `json:"mid"`
	// code indicating success or failure of the request
	Code int `json:"code,string"`
	// message with more details about the result
	Message string `json:"message"`
	// timestamp
	Timestamp int64 `json:"timestamp,string"`
	// Data
	Data interface{} `json:"data"`
}

func (r *Response) Marshal() (data []byte, err error) {
	return jsoniter.Marshal(r)
}

type PushMessageResponse struct {
	MessageID int64 `json:"message_id,string"`
	Sequence  int64 `json:"sequence,string"`
}

func NewResponse(err error, mid string, timestamp int64, data interface{}) *Response {
	code := ecode.Cause(err)
	if data == nil {
		data = struct{}{}
	}
	return &Response{
		MID:       mid,
		Code:      code.Code(),
		Message:   code.Message(),
		Timestamp: timestamp,
		Data:      data,
	}
}

type ServerMessage struct {
	Data      []byte
	Timestamp int64
}

func NoErr(mid string, timestamp int64, data interface{}) []byte {
	resp, _ := NewResponse(ecode.OK, mid, timestamp, data).Marshal()
	return resp
}

// NoErrShutdown means user was disconnected from topic because system shutdown is in progress.
func NoErrShutdown() []byte {
	resp, _ := NewResponse(ecode.ResetContent.ResetMessage("server shutdown"), "", 0, nil).Marshal()
	return resp
}

// ErrMalformed bad request.
func ErrBadRequest(mid string, timestamp int64) []byte {
	resp, _ := NewResponse(ecode.ErrBadRequest, mid, timestamp, nil).Marshal()
	return resp
}

// ErrMalformed request malformed.
func ErrMalformed(mid string, timestamp int64) []byte {
	resp, _ := NewResponse(ecode.ErrBadRequest.ResetMessage("malformed"), mid, timestamp, nil).Marshal()
	return resp
}

// ErrAuthRequired authentication required  - user must authenticate first.
func ErrAuthRequired(mid string, timestamp int64) []byte {
	resp, _ := NewResponse(ecode.ErrUnauthorized.ResetMessage("authentication required"), mid, timestamp, nil).Marshal()
	return resp
}

// ErrAuthFailed authentication failed.
func ErrAuthFailed(mid string, timestamp int64) []byte {
	resp, _ := NewResponse(ecode.ErrUnauthorized.ResetMessage("authentication failed"), mid, timestamp, nil).Marshal()
	return resp
}

// ErrVersionNotSupported invalid (too low) protocol version.
func ErrVersionNotSupported(mid string, timestamp int64) []byte {
	resp, _ := NewResponse(ecode.ErrBadRequest.ResetMessage("version not supported"), mid, timestamp, nil).Marshal()
	return resp
}

// ErrInternalServer database or other server error.
func ErrInternalServer(mid string, timestamp int64, message string) []byte {
	resp, _ := NewResponse(ecode.ErrInternalServer.ResetMessage(message), mid, timestamp, nil).Marshal()
	return resp
}
