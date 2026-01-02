package common

import "errors"

var WIFInputError = errors.New("the format of the private key WIF is invalid")
var ServerError = errors.New("internal server error")
