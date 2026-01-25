package main

import "errors"

// IdentityError defines the error types for identity service
type IdentityError error

var ErrUserNotFound IdentityError = errors.New("User not found")
var ErrInvalidSignature IdentityError = errors.New("Invalid signature")
