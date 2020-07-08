package session

import "outgoing/x/types"

type ServerMessage struct {
	Data      []byte
	Timestamp int64
}

type serverMessage struct {
	MID       string
	RouteTo   string
	AsUser    string
	Timestamp int64
	sess      *Session
	SkipSid   string
	uid       types.Uid
}
