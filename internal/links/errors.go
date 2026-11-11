package links

import (
	"errors"
)

var ErrConnection = errors.New("connection is interrupted")
var ErrEmpty = errors.New("data can not be empty")
var ErrEmptyBaseURL = errors.New("base URL can not be empty")
var ErrUnsupportScheme = errors.New("URL must use the http or https scheme")
var ErrEmptyHostname = errors.New("hostname can not be empty")
var ErrEmptyURL = errors.New("URL can not be empty")
var ErrInvalidID = errors.New("ID must be more than 0")
var ErrEmptyShortName = errors.New("short name can not be empty")
var ErrNotFound = errors.New("link not found")
var ErrValidate = errors.New("validation error")
