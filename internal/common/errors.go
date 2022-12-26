package common

import (
	"errors"
)

var (
	ErrInvalidDockerClient = errors.New("invalid docker client")
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
)
