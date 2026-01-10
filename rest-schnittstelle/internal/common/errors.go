package common

import "errors"

var ErrWIFInput = errors.New("the format of the private key WIF is invalid")
var ErrServer = errors.New("internal server error")
var ErrInvalidAddress = errors.New("invalid VSAddress format")

type AssetError struct {
	Message string
}

func (e *AssetError) Error() string {
	return e.Message
}
