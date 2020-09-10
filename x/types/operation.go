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
	OperationConnect
	OperationPush
	OperationNotification
	OperationBroadcast
)

// String implements Stringer interface: gets human-readable name for a numeric operation.
func (o Operation) String() string {
	s, err := o.MarshalText()
	if err != nil {
		return "unknown"
	}
	return string(s)
}

// ParseOperation parses operation from a string.
func ParseOperation(name string) Operation {
	var o Operation
	_ = o.UnmarshalText([]byte(name))
	return o
}

// MarshalText converts Operation to a slice of bytes with the name of the operation.
func (o Operation) MarshalText() ([]byte, error) {
	switch o {
	case OperationHandshake:
		return []byte("handshake"), nil
	case OperationHeartbeat:
		return []byte("heartbeat"), nil
	case OperationConnect:
		return []byte("connect"), nil
	case OperationPush:
		return []byte("push"), nil
	case OperationNotification:
		return []byte("notification"), nil
	case OperationBroadcast:
		return []byte("broadcast"), nil
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
	case "connect":
		*o = OperationConnect
	case "push":
		*o = OperationPush
	case "notification":
		*o = OperationNotification
	case "broadcast":
		*o = OperationBroadcast
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
