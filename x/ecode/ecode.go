package ecode

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

var (
	_codes = map[int]string{}
)

// New new a ecode.Codes by int value.
// NOTE: ecode must unique in global, the New will check repeat and then panic.
func New(e int, m string) Code {
	if e <= 0 {
		panic("business ecode must greater than zero")
	}
	return add(e, m)
}

func add(e int, m string) Code {
	if _, ok := _codes[e]; ok {
		panic(fmt.Sprintf("ecode: %d already exist", e))
	}
	_codes[e] = m
	return Code{code: e, message: m}
}

type Code struct {
	code    int
	message string
}

func (c Code) Error() string {
	return c.Message()
}

// Code return error code
func (c Code) Code() int { return c.code }

// Message return error message
func (c Code) Message() string {
	if c.message != "" {
		return c.message
	}
	if m, ok := _codes[c.code]; ok {
		return m
	}
	return strconv.Itoa(c.code)
}

// ResetMessage reset error message
func (c Code) ResetMessage(message string) error {
	c.message = message
	return c
}

// Details return details.
func (c Code) Details() []interface{} { return nil }

// Equal for compatible.
// Deprecated: please use ecode.EqualError.
func (c Code) Equal(err error) bool { return EqualError(c, err) }

// Codes ecode error interface which has a code & message.
type Coder interface {
	// sometimes Error return Code in string form
	// NOTE: don't use Error in monitor report even it also work for now
	Error() string
	// Code get error code.
	Code() int
	// Message get code message.
	Message() string
	// Message reset code message.
	ResetMessage(message string) error
	//Detail get error detail,it may be nil.
	Details() []interface{}
	// Equal for compatible.
	// Deprecated: please use ecode.EqualError.
	Equal(error) bool
}

// Cause cause from error to ecode.
func Cause(e error) Coder {
	if e == nil {
		return OK
	}

	c, ok := errors.Cause(e).(Coder)
	if ok {
		return c
	}
	if c == nil {
		return ErrInternalServer
	}
	return Code{code: c.Code(), message: c.Message()}
}

func EqualError(code Coder, err error) bool {
	return code.Code() == Cause(err).Code()
}
