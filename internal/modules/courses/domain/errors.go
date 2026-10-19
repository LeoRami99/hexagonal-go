package domain

import "errors"

var (
	ErrCourseNotFound      = errors.New("course not found")
	ErrCourseAlreadyExists = errors.New("course already exists")
	ErrInvalidInput        = errors.New("invalid course input")
)
