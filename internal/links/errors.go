package links 

import (
	"errors"
)

var ErrNotFound = errors.New("object not found")
var ConnectionErr = errors.New("connection is interrupted")
var EmptyErr = errors.New("data can not be empty")
var EmptyBaseURLErr = errors.New("base URL can not be empty")
var UnsupportedSchemeErr = errors.New("URL must use the http or https scheme")
var EmptyHostnameErr = errors.New("hostname can not be empty")
var EmptyURLErr = errors.New("URL can not be empty")
var InvalidIDErr = errors.New("ID must be more than 0")
var EmptyShortNameErr = errors.New("short name can not be empty")
var ErrNotFound = errors.New("link not found")
var ValidationErr = errors.New("validation error")
