package usecase

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAdd = errors.New("Failed to add user")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidToken = errors.New("token is invalid")
