package apperrors

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrServer UserErr = "server error"

	ErrEmailVerify          UserErr = "please verify email"
	ErrUserNotRegister      UserErr = "the user is not registered"
	ErrInvalidPassword      UserErr = "invalid password"
	ErrUserAlreadyActivated UserErr = "user already activated"

	// jwt errors
	ErrEmailTokenExpired       UserErr = "email verification token is expired, request a new token"
	ErrInvalidEmailVerifyToken UserErr = "email verification token is invalid"

	ErrInvalidRefreshToken UserErr = "refresh token is invalid"
	ErrRefreshTokenExpired UserErr = "refresh token is expired"

	ErrAccessTokenExpired UserErr = "access token is expired"
	ErrInvalidAccessToken UserErr = "access token is invalid"

	ErrPleaseLogin UserErr = "to use this feature, please login"
	// prodct
	ErrProductNotFound UserErr = "product not found"
)

func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	var userErr UserError
	if errors.As(err, &userErr) {
		return err
	}
	return fmt.Errorf("%s -> %w", op, err)
}

type UserErr string

func (e UserErr) Error() string {
	return string(e)
}

func (e UserErr) UserError() {}

type UserError interface {
	error
	UserError()
}

type ErrLoginExists struct {
	Username string
	Email    string
}

func (e *ErrLoginExists) Error() string {
	var str strings.Builder
	if e.Username != "" {
		str.WriteString(fmt.Sprintf("username: %s, exists; ", e.Username))
	}
	if e.Email != "" {
		str.WriteString(fmt.Sprintf("email: %s, exists; ", e.Email))
	}
	return strings.TrimSpace(str.String())
}
func (e *ErrLoginExists) UserError() {}

type ValidationError struct {
	Field string
	Err   error
}

func (e *ValidationError) Error() string {
	return e.Err.Error()
}
func (e *ValidationError) UserError() {}

type ValidationErrors []ValidationError

func (e *ValidationErrors) UserError() {}
func (e *ValidationErrors) Add(field, err string) {
	*e = append(*e, ValidationError{
		Field: field,
		Err:   errors.New(err),
	})

}
func (e *ValidationErrors) Error() string {
	var err strings.Builder

	for _, v := range *e {
		err.WriteString(v.Field)
		err.WriteString(": ")
		err.WriteString(v.Err.Error() + "; ")
	}
	return strings.TrimSpace(err.String())
}
