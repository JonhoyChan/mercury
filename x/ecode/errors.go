package ecode

import (
	"fmt"
	"github.com/pkg/errors"
)

func NewError(message string) error {
	return errors.New(message)
}

func NewErrorf(format string, a ...interface{}) error {
	return errors.New(fmt.Sprintf(format, a))
}

func Wrap(e error, message string) error {
	return errors.Wrap(e, message)
}

func Wrapf(e error, format string, args ...interface{}) error {
	return errors.Wrapf(e, format, args)
}

func Unwrap(e error) error {
	return errors.Unwrap(e)
}
