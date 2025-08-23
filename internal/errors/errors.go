package errors

import "errors"

var ErrUnauthorized = errors.New("user is not authorized")
var ErrForbidden = errors.New("operation is forbidden for user")
