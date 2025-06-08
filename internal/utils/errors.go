package utils

import "errors"

var (
	ErrInvalidTokenClaims = errors.New("invalid token claims")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenMalformed     = errors.New("token malformed")
	ErrSignatureInvalid   = errors.New("signature invalid")
)
