package storage

import (
	// embedded
	"errors"
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURlExists = errors.New("url already exists")
)