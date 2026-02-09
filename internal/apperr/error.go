package apperr

import "fmt"

type Kind string

const (
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindInvalid      Kind = "invalid"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindRateLimited  Kind = "rate_limited"
	KindUnavailable  Kind = "unavailable"
	KindExternal     Kind = "external"
	KindInternal     Kind = "internal"
)

type Error struct {
	Kind Kind
	Op   string
	Err  error
	Msg  string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Msg != "" {
		return e.Msg
	}
	if e.Op == "" {
		if e.Err != nil {
			return e.Err.Error()
		}
		return string(e.Kind)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Op
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok || t.Kind == "" {
		return false
	}
	return e.Kind == t.Kind
}

var (
	ErrNotFound     = &Error{Kind: KindNotFound}
	ErrConflict     = &Error{Kind: KindConflict}
	ErrInvalid      = &Error{Kind: KindInvalid}
	ErrUnauthorized = &Error{Kind: KindUnauthorized}
	ErrForbidden    = &Error{Kind: KindForbidden}
	ErrRateLimited  = &Error{Kind: KindRateLimited}
	ErrUnavailable  = &Error{Kind: KindUnavailable}
	ErrExternal     = &Error{Kind: KindExternal}
	ErrInternal     = &Error{Kind: KindInternal}
)

func Wrap(kind Kind, op string, err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{Kind: kind, Op: op, Err: err}
}
