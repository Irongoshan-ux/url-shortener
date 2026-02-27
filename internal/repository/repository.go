package repository

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("url not found")
	ErrAlreadyExists = errors.New("url already exists")
	ErrConflict      = errors.New("original url already shortened")
)
