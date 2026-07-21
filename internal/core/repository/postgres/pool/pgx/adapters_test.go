package pgx

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool/pgx/mocks"
	"go.uber.org/mock/gomock"
)

func TestAdapters_ErrRowsMapping(t *testing.T) {
	tests := []struct {
		name string
		call func(t *testing.T, ctrl *gomock.Controller) error
	}{
		{
			name: "pgxRow.Scan maps pgx.ErrNoRows to pool.ErrNoRows",
			call: func(t *testing.T, ctrl *gomock.Controller) error {
				mockRow := mocks.NewMockRow(ctrl)
				mockRow.EXPECT().
					Scan(gomock.Any()).
					Return(pgx.ErrNoRows).
					Times(1)

				var dest string
				return pgxRow{mockRow}.Scan(&dest)
			},
		},
		{
			name: "pgxRows.Scan maps pgx.ErrNoRows to pool.ErrNoRows",
			call: func(t *testing.T, ctrl *gomock.Controller) error {
				mockRows := mocks.NewMockRows(ctrl)
				mockRows.EXPECT().
					Scan(gomock.Any()).
					Return(pgx.ErrNoRows).
					Times(1)

				var dest string
				return pgxRows{mockRows}.Scan(&dest)
			},
		},
		{
			name: "pgxRows.Err maps pgx.ErrNoRows to pool.ErrNoRows",
			call: func(t *testing.T, ctrl *gomock.Controller) error {
				mockRows := mocks.NewMockRows(ctrl)
				mockRows.EXPECT().
					Err().
					Return(pgx.ErrNoRows).
					Times(1)

				return pgxRows{mockRows}.Err()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			err := tt.call(t, ctrl)

			assert.ErrorIs(t, err, pool.ErrNoRows)
		})
	}
}

func TestMapErrors(t *testing.T) {
	testCases := []struct {
		name        string
		actualErr   error
		expectedErr error
	}{
		{
			name:        "maps pgx.ErrNoRows to pool.ErrNoRows",
			actualErr:   pgx.ErrNoRows,
			expectedErr: pool.ErrNoRows,
		},
		{
			name:        "maps pgconn.PgErr 23503 to pool.ErrViolatesForeignKey",
			actualErr:   &pgconn.PgError{Code: "23503"},
			expectedErr: pool.ErrViolatesForeignKey,
		}, {
			name:        "maps pgconn.PgErr 23505 to pool.ErrUniqueViolation",
			actualErr:   &pgconn.PgError{Code: "23505"},
			expectedErr: pool.ErrUniqueViolation,
		},
		{
			name:        "maps unmapped pgconn.PgErr code to pool.ErrUnknown",
			actualErr:   &pgconn.PgError{Code: "42P01"},
			expectedErr: pool.ErrUnknown,
		},
		{
			name:        "maps another errors to pool.ErrUnkown",
			actualErr:   errors.New("some error"),
			expectedErr: pool.ErrUnknown,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := mapErrors(tt.actualErr)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
