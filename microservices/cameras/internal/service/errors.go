package service

import "errors"

var (
	ErrCameraNotFound error = errors.New("Camera not found")
	ErrSizeZero       error = errors.New("Size 0")
)
