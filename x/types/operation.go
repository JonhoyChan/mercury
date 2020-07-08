package types

import (
	"errors"
	"strings"
)

// Operation is the type for request.
type Operation int8

// Operations
const (
	OperationUnknown Operation = iota
	OperationHandshake
	OperationHeartbeat
	OperationAuthenticate
	OperationSync
	OperationPublish
	OperationSubscribe
	OperationUnsubscribe
)

// String implements Stringer interface: gets human-readable name for a numeric operation.
func (o Operation) String() string {
	s, err := o.MarshalText()
	if err != nil {
		return "unknown"
	}
	return string(s)
}

// ParseAuthLevel parses operation from a string.
func ParseOperation(name string) Operation {
	switch strings.ToLower(name) {
	case "handshake":
		return OperationHandshake
	case "heartbeat":
		return OperationHeartbeat
	case "authenticate":
		return OperationAuthenticate
	case "sync":
		return OperationSync
	case "publish":
		return OperationPublish
	case "subscribe":
		return OperationSubscribe
	case "unsubscribe":
		return OperationUnsubscribe
	default:
		return OperationUnknown
	}
}

// MarshalText converts Operation to a slice of bytes with the name of the operation.
func (o Operation) MarshalText() ([]byte, error) {
	switch o {
	case OperationHandshake:
		return []byte("handshake"), nil
	case OperationHeartbeat:
		return []byte("heartbeat"), nil
	case OperationAuthenticate:
		return []byte("authenticate"), nil
	case OperationSync:
		return []byte("sync"), nil
	case OperationPublish:
		return []byte("publish"), nil
	case OperationSubscribe:
		return []byte("subscribe"), nil
	case OperationUnsubscribe:
		return []byte("unsubscribe"), nil
	default:
		return []byte("unknown"), nil
	}
}

// UnmarshalText parses Operation from a string.
func (o *Operation) UnmarshalText(b []byte) error {
	switch strings.ToLower(string(b)) {
	case "handshake":
		*o = OperationHandshake
	case "heartbeat":
		*o = OperationHeartbeat
	case "authenticate":
		*o = OperationAuthenticate
	case "sync":
		*o = OperationSync
	case "publish":
		*o = OperationPublish
	case "subscribe":
		*o = OperationSubscribe
	case "unsubscribe":
		*o = OperationUnsubscribe
	default:
		*o = OperationUnknown
	}
	return nil
}

// MarshalJSON converts Operation to a quoted string.
func (o Operation) MarshalJSON() ([]byte, error) {
	res, err := o.MarshalText()
	if err != nil {
		return nil, err
	}

	return append(append([]byte{'"'}, res...), '"'), nil
}

// UnmarshalJSON reads Operation from a quoted string.
func (o *Operation) UnmarshalJSON(b []byte) error {
	if b[0] != '"' || b[len(b)-1] != '"' {
		return errors.New("syntax error")
	}

	return o.UnmarshalText(b[1 : len(b)-1])
}
