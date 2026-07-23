package pgx

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"
)

//go:generate mockgen -destination=mocks/mock_pgx.go -package=mocks github.com/jackc/pgx/v5 Row,Rows
type pgxRows struct {
	pgx.Rows
}

func (r pgxRows) Scan(dest ...any) error {
	err := r.Rows.Scan(dest...)

	if err != nil {
		return mapErrors(err)
	}

	return nil
}

func (r pgxRows) Err() error {
	err := r.Rows.Err()

	if err != nil {
		return mapErrors(err)
	}

	return nil
}

type pgxRow struct {
	pgx.Row
}

func (r pgxRow) Scan(dest ...any) error {
	err := r.Row.Scan(dest...)

	if err != nil {
		return mapErrors(err)
	}

	return nil
}

type commandTag struct {
	pgconn.CommandTag
}

func mapErrors(err error) error {
	const (
		pgxViolatesForeignKeyErrCode = "23503"
		pgxUniqueViolationErrCode    = "23505"
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return pool.ErrNoRows
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case pgxViolatesForeignKeyErrCode:
			return fmt.Errorf("%v: %w", err, pool.ErrViolatesForeignKey)
		case pgxUniqueViolationErrCode:
			return fmt.Errorf("%v: %w", err, pool.ErrUniqueViolation)
		}
	}

	return fmt.Errorf("%v: %w", err, pool.ErrUnknown)
}
