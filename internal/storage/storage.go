package storage

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURlExists = errors.New("url already exists")
)