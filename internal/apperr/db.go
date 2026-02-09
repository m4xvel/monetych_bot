package apperr

import "fmt"

const (
	DBCodeOrderAlreadyProcessed = "order_already_processed"
)

type DBError struct {
	Op         string
	Kind       Kind
	Code       string
	Constraint string
	Err        error
}

func (e *DBError) Error() string {
	if e == nil {
		return ""
	}
	base := "db error"
	if e.Op != "" {
		base = e.Op
	}
	if e.Code != "" {
		base = fmt.Sprintf("%s: %s", base, e.Code)
	}
	if e.Constraint != "" {
		base = fmt.Sprintf("%s (constraint %s)", base, e.Constraint)
	}
	if e.Err != nil {
		base = fmt.Sprintf("%s: %v", base, e.Err)
	}
	return base
}

func (e *DBError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *DBError) Is(target error) bool {
	switch t := target.(type) {
	case *Error:
		if t.Kind == "" {
			return false
		}
		return e.Kind == t.Kind
	case *DBError:
		if t.Kind != "" && e.Kind != t.Kind {
			return false
		}
		if t.Code != "" && e.Code != t.Code {
			return false
		}
		if t.Constraint != "" && e.Constraint != t.Constraint {
			return false
		}
		return true
	default:
		return false
	}
}
