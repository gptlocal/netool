package rpc

import (
	"errors"
	"reflect"
)

const (
	logRegisterError = true
)

type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var (
	ErrShutdown = errors.New("connection is shut down")

	// Precompute the reflect type for error. Can't use error directly because Typeof takes an empty interface value. This is annoying.
	typeOfError = reflect.TypeOf((*error)(nil)).Elem()
)
