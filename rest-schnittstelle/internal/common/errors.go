package common

import "errors"

var ErrWIFInput = errors.New("the format of the private key WIF is invalid")
var ErrServer = errors.New("internal server error")
