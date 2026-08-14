package postgres

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/geo"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool/mocks"
	"go.uber.org/mock/gomock"
)

var (
	placeholderRe = regexp.MustCompile(`\$(\d+)`)
	columnsRe     = regexp.MustCompile(`INSERT INTO analytics\.clicks \(([^)]+)\)`)
)

func parseColumns(t *testing.T, query string) []string {
	t.Helper()

	matches := columnsRe.FindStringSubmatch(query)
	require.Len(
		t,
		matches,
		2,
		"query must contain INSERT INTO analytics.clicks (...) column list",
	)

	rawColumns := strings.Split(matches[1], ",")
	columns := make([]string, 0, len(rawColumns))

	for _, column := range rawColumns {
		columns = append(columns, strings.TrimSpace(column))
	}

	return columns
}

func assertContiguousPlaceholders(t *testing.T, query string, expectedCount int) {
	t.Helper()

	matches := placeholderRe.FindAllStringSubmatch(query, -1)
	require.Len(t, matches, expectedCount)

	for i, match := range matches {
		got, err := strconv.Atoi(match[1])
		require.NoError(t, err)

		assert.Equal(
			t,
			i+1,
			got,
			"placeholder at position %d must be $%d, got $%d",
			i,
			i+1,
			got,
		)
	}
}

func initTest(t *testing.T) (*Repository, *mocks.MockPool) {
	t.Helper()

	ctrl := gomock.NewController(t)
	poolMock := mocks.NewMockPool(ctrl)

	poolMock.EXPECT().
		OpTimeout().
		Return(5 * time.Second).
		Times(1)

	return NewRepository(poolMock), poolMock
}

func TestSaveClicks(t *testing.T) {
	inputEvents := []events.ClickEvent{
		events.NewClickEvent(
			"shortCode-0",
			geo.Geo{
				Country: "US",
				City:    "New York",
			},
		),
		events.NewClickEvent(
			"shortCode-1",
			geo.Geo{
				Country: "DE",
				City:    "Berlin",
			},
		),
		events.NewClickEvent(
			"shortCode-2",
			geo.Geo{
				Country: "JP",
				City:    "Tokyo",
			},
		),
	}

	expectedColumns := []string{
		"id",
		"short_code",
		"country",
		"city",
		"clicked_at",
	}

	expectedArgs := make([]any, 0, len(inputEvents)*len(expectedColumns))

	for _, event := range inputEvents {
		expectedArgs = append(
			expectedArgs,
			event.ID,
			event.ShortCode,
			event.CountryCode,
			event.City,
			event.ClickedAt,
		)
	}

	testCases := []struct {
		name         string
		rowsAffected int64
		execErr      error
		check        func(t *testing.T, err error)
	}{
		{
			name:         "success, all rows inserted",
			rowsAffected: int64(len(inputEvents)),
			check: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:         "some rows were duplicates",
			rowsAffected: int64(len(inputEvents)) - 1,
			check: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperrors.ErrConflict)
			},
		},
		{
			name:    "exec fails",
			execErr: errors.New("connection reset"),
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.NotErrorIs(t, err, apperrors.ErrConflict)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repository, poolMock := initTest(t)

			var (
				capturedQuery string
				capturedArgs  []any
			)

			poolMock.EXPECT().
				Exec(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(
					func(
						_ context.Context,
						query string,
						args ...any,
					) (pool.CommandTag, error) {
						capturedQuery = query
						capturedArgs = args

						if tc.execErr != nil {
							return nil, tc.execErr
						}

						commandTag := mocks.NewMockCommandTag(gomock.NewController(t))
						commandTag.EXPECT().
							RowsAffected().
							Return(tc.rowsAffected).
							Times(1)

						return commandTag, nil
					},
				).
				Times(1)

			err := repository.SaveClicks(context.Background(), inputEvents)
			tc.check(t, err)

			assert.Equal(t, expectedColumns, parseColumns(t, capturedQuery))

			require.Len(
				t,
				capturedArgs,
				len(inputEvents)*len(expectedColumns),
			)

			assert.Equal(t, expectedArgs, capturedArgs)

			assertContiguousPlaceholders(
				t,
				capturedQuery,
				len(inputEvents)*len(expectedColumns),
			)
		})
	}
}
