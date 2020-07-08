package types

import (
	"errors"
	"strings"
)

// AuthLevel is the type for authentication levels.
type AuthLevel int

// Authentication levels
const (
	// AuthLevelNone is undefined/not authenticated
	AuthLevelNone AuthLevel = iota * 10
	// AuthLevelAnon is anonymous user/light authentication
	AuthLevelAnon
	// AuthLevelAuth is fully authenticated user
	AuthLevelAuth
	// AuthLevelRoot is a superuser (currently unused)
	AuthLevelRoot
)

// String implements Stringer interface: gets human-readable name for a numeric authentication level.
func (a AuthLevel) String() string {
	s, err := a.MarshalText()
	if err != nil {
		return "unknown"
	}
	return string(s)
}

// ParseAuthLevel parses authentication level from a string.
func ParseAuthLevel(name string) AuthLevel {
	switch strings.ToLower(name) {
	case "anon":
		return AuthLevelAnon
	case "auth":
		return AuthLevelAuth
	case "root":
		return AuthLevelRoot
	default:
		return AuthLevelNone
	}
}

// MarshalText converts Level to a slice of bytes with the name of the level.
func (a AuthLevel) MarshalText() ([]byte, error) {
	switch a {
	case AuthLevelNone:
		return []byte(""), nil
	case AuthLevelAnon:
		return []byte("anon"), nil
	case AuthLevelAuth:
		return []byte("auth"), nil
	case AuthLevelRoot:
		return []byte("root"), nil
	default:
		return nil, errors.New("auth.Level: invalid level value")
	}
}

// UnmarshalText parses authentication level from a string.
func (a *AuthLevel) UnmarshalText(b []byte) error {
	switch strings.ToLower(string(b)) {
	case "":
		*a = AuthLevelNone
		return nil
	case "anon":
		*a = AuthLevelAnon
		return nil
	case "auth":
		*a = AuthLevelAuth
		return nil
	case "root":
		*a = AuthLevelRoot
		return nil
	default:
		return errors.New("auth.Level: unrecognized")
	}
}

// MarshalJSON converts Level to a quoted string.
func (a AuthLevel) MarshalJSON() ([]byte, error) {
	res, err := a.MarshalText()
	if err != nil {
		return nil, err
	}

	return append(append([]byte{'"'}, res...), '"'), nil
}

// UnmarshalJSON reads Level from a quoted string.
func (a *AuthLevel) UnmarshalJSON(b []byte) error {
	if b[0] != '"' || b[len(b)-1] != '"' {
		return errors.New("syntax error")
	}

	return a.UnmarshalText(b[1 : len(b)-1])
}
