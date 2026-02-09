package postgres

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/m4xvel/monetych_bot/internal/apperr"
)

const (
	pgErrUniqueViolation      = "23505"
	pgErrForeignKeyViolation  = "23503"
	pgErrNotNullViolation     = "23502"
	pgErrCheckViolation       = "23514"
	pgErrSerializationFailure = "40001"
	pgErrDeadlockDetected     = "40P01"
	pgErrLockNotAvailable     = "55P03"
)

func dbErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return &apperr.DBError{Op: op, Kind: apperr.KindNotFound, Err: err}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		kind := apperr.KindInternal
		switch pgErr.Code {
		case pgErrUniqueViolation:
			kind = apperr.KindConflict
		case pgErrForeignKeyViolation,
			pgErrNotNullViolation,
			pgErrCheckViolation:
			kind = apperr.KindInvalid
		case pgErrSerializationFailure,
			pgErrDeadlockDetected,
			pgErrLockNotAvailable:
			kind = apperr.KindUnavailable
		}
		return &apperr.DBError{
			Op:         op,
			Kind:       kind,
			Code:       pgErr.Code,
			Constraint: pgErr.ConstraintName,
			Err:        err,
		}
	}

	return &apperr.DBError{Op: op, Kind: apperr.KindInternal, Err: err}
}

func dbErrKind(op string, kind apperr.Kind, err error) error {
	return &apperr.DBError{Op: op, Kind: kind, Err: err}
}

func dbErrCode(op string, kind apperr.Kind, code string, err error) error {
	return &apperr.DBError{Op: op, Kind: kind, Code: code, Err: err}
}
