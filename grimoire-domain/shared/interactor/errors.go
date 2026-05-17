package interactor

import "errors"

var ErrUnauthorized = errors.New("interactor: caller is not authorized")