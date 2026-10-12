//go:build !cgo

package coccoc

import (
	"errors"
)

var errCGONotAvailable = errors.New("CGO không khả dụng trong bản build này (CGO_ENABLED=0 hoặc thiếu C++ compiler)")

func newCGOEngine(opts Options) (Tokenizer, error) {
	return nil, errCGONotAvailable
}
