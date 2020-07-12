package session

import (
	"encoding/json"
	"fmt"
	"outgoing/app/gateway/chat/api"
	"outgoing/x/ecode"
)

type ServerMessage struct {
	Data      []byte
	Timestamp int64
}

func NoErr(mid string, timestamp int64) []byte {
	//data, _ := api.NewResponse(ecode.OK, mid, timestamp).Marshal()
	data, err := json.Marshal(api.NewResponse(ecode.OK, mid, timestamp))
	if err != nil {
		fmt.Println(err.Error())
	}
	return data
}

// NoErrShutdown means user was disconnected from topic because system shutdown is in progress.
func NoErrShutdown() []byte {
	data, _ := api.NewResponse(ecode.ResetContent.ResetMessage("server shutdown"), "", 0).Marshal()
	return data
}

// ErrMalformed request malformed.
func ErrMalformed(mid string, timestamp int64) []byte {
	//data, _ := api.NewResponse(ecode.ErrBadRequest.ResetMessage("malformed"), mid, timestamp).Marshal()
	data, _ := json.Marshal(api.NewResponse(ecode.ErrBadRequest.ResetMessage("malformed"), mid, timestamp))
	return data
}

// ErrAuthRequired authentication required  - user must authenticate first.
func ErrAuthRequired(mid string, timestamp int64) []byte {
	//data, _ := api.NewResponse(ecode.ErrUnauthorized.ResetMessage("authentication required"), mid, timestamp).Marshal()
	data, _ := json.Marshal(api.NewResponse(ecode.ErrUnauthorized.ResetMessage("authentication required"), mid, timestamp))
	return data
}

// ErrAuthFailed authentication failed.
func ErrAuthFailed(mid string, timestamp int64) []byte {
	//data, _ := api.NewResponse(ecode.ErrUnauthorized.ResetMessage("authentication failed"), mid, timestamp).Marshal()
	data, _ := json.Marshal(api.NewResponse(ecode.ErrUnauthorized.ResetMessage("authentication failed"), mid, timestamp))
	return data
}

// ErrVersionNotSupported invalid (too low) protocol version.
func ErrVersionNotSupported(mid string, timestamp int64) []byte {
	data, _ := api.NewResponse(ecode.ErrBadRequest.ResetMessage("version not supported"), mid, timestamp).Marshal()
	return data
}

// ErrInternalServer database or other server error.
func ErrInternalServer(mid string, timestamp int64) []byte {
	data, _ := api.NewResponse(ecode.ErrInternalServer, mid, timestamp).Marshal()
	return data
}
