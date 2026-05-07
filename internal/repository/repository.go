// Package repository provides in-memory, file-backed, and PostgreSQL implementations of service.Repository.
package repository

import (
	"errors"
)

var (
	// ErrNotFound means no row exists for the requested short or original URL.
	ErrNotFound = errors.New("url not found")
	// ErrAlreadyExists means the generated short id is already taken (retry with a new id).
	ErrAlreadyExists = errors.New("url already exists")
	// ErrConflict means a non-deleted mapping for the original URL already exists.
	ErrConflict = errors.New("original url already shortened")
)
