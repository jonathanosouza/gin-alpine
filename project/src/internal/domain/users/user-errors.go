// Package users
package users

import "errors"

var (
	ErrUserEmailAlreadyExists = errors.New("user with this email already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidRole            = errors.New("invalid role")
	ErrRoleNotAllowed         = errors.New("role not allowed")
)
