package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type Kind string

const (
	Validation   Kind = "validation_error"
	NotFound     Kind = "not_found"
	Conflict     Kind = "conflict"
	Unauthorized Kind = "unauthorized"
	Forbidden    Kind = "forbidden"
	RateLimited  Kind = "rate_limited"
	Timeout      Kind = "timeout"
	Unavailable  Kind = "unavailable"
	Internal     Kind = "internal_error"
)

type AppError struct {
	Kind   Kind
	Op     string            // optional: operation / usecase (e.g. "users.create")
	Msg    string            // safe message (no PII)
	Err    error             // wrapped error (stack cause)
	Fields map[string]string // safe fields (no PII)
}

func (e *AppError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %s", e.Op, e.Msg)
	}
	return e.Msg
}
func (e *AppError) Unwrap() error { return e.Err }

func E(kind Kind, msg string, cause error) *AppError {
	return &AppError{Kind: kind, Msg: msg, Err: cause}
}
func Op(op string, err error) *AppError {
	if err == nil {
		return nil
	}
	var ae *AppError
	if errors.As(err, &ae) {
		ae.Op = op
		return ae
	}
	return &AppError{Kind: Internal, Op: op, Msg: err.Error(), Err: err}
}
func WithField(err error, k, v string) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		if ae.Fields == nil {
			ae.Fields = map[string]string{}
		}
		ae.Fields[k] = v
		return ae
	}
	return &AppError{Kind: Internal, Msg: "wrapped non-app error", Err: err, Fields: map[string]string{k: v}}
}

func IsKind(err error, k Kind) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Kind == k
	}
	return false
}

func StatusFor(err error) int {
	var ae *AppError
	if errors.As(err, &ae) {
		switch ae.Kind {
		case Validation:
			return http.StatusBadRequest
		case NotFound:
			return http.StatusNotFound
		case Conflict:
			return http.StatusConflict
		case Unauthorized:
			return http.StatusUnauthorized
		case Forbidden:
			return http.StatusForbidden
		case RateLimited:
			return http.StatusTooManyRequests
		case Timeout:
			return http.StatusGatewayTimeout
		case Unavailable:
			return http.StatusServiceUnavailable
		case Internal:
			fallthrough
		default:
			return http.StatusInternalServerError
		}
	}
	// unknown → 500
	return http.StatusInternalServerError
}
