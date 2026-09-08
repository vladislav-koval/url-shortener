package postgres

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/domain"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
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
	inputClicks := []domain.Click{
		{
			ID:        uuid.New(),
			ShortCode: "shortCode-0",
			ClickedAt: time.Now(),
			Country:   "US",
			City:      "New York",

			DeviceType: "desktop",
			OS:         "Windows",
			Browser:    "Chrome",

			Referer: "http://referer-0.com",
		},
		{
			ID:        uuid.New(),
			ShortCode: "shortCode-1",
			ClickedAt: time.Now(),
			Country:   "DE",
			City:      "Berlin",

			DeviceType: "mobile",
			OS:         "iOS",
			Browser:    "Safari",

			Referer: "http://referer-1.com",
		},
		{
			ID:        uuid.New(),
			ShortCode: "shortCode-2",
			ClickedAt: time.Now(),
			Country:   "JP",
			City:      "Tokyo",

			DeviceType: "tablet",
			OS:         "Android",
			Browser:    "Firefox",

			Referer: "http://referer-2.com",
		},
	}

	expectedColumns := []string{
		"id",
		"short_code",
		"country",
		"city",
		"clicked_at",
		"device_type",
		"os",
		"browser",
		"referer",
	}

	expectedArgs := make([]any, 0, len(inputClicks)*len(expectedColumns))

	for _, click := range inputClicks {
		expectedArgs = append(
			expectedArgs,
			click.ID,
			click.ShortCode,
			click.Country,
			click.City,
			click.ClickedAt,
			click.DeviceType,
			click.OS,
			click.Browser,
			click.Referer,
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
			rowsAffected: int64(len(inputClicks)),
			check: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:         "some rows were duplicates",
			rowsAffected: int64(len(inputClicks)) - 1,
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

			err := repository.SaveClicks(context.Background(), inputClicks)
			tc.check(t, err)

			assert.Equal(t, expectedColumns, parseColumns(t, capturedQuery))

			require.Len(
				t,
				capturedArgs,
				len(inputClicks)*len(expectedColumns),
			)

			assert.Equal(t, expectedArgs, capturedArgs)

			assertContiguousPlaceholders(
				t,
				capturedQuery,
				len(inputClicks)*len(expectedColumns),
			)

			// Regression: fmt.Sprintf-собранный список плейсхолдеров однажды получил
			// лишнюю запятую перед закрывающей скобкой каждого VALUES(...) — плейсхолдеры
			// $1..$n при этом оставались сплошными (assertContiguousPlaceholders такое не
			// ловит, запятая — не плейсхолдер), а запрос падал на реальном Postgres с
			// синтаксической ошибкой. Проверено вживую: "VALUES (1,2,...,9,)" — syntax
			// error at or near ")".
			assert.NotContains(t, capturedQuery, ",)", "trailing comma before closing paren makes the query invalid SQL")
		})
	}
}

func TestSaveClicks_EmptyInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	poolMock := mocks.NewMockPool(ctrl)
	repository := NewRepository(poolMock)

	// Пустой VALUES-список ("INSERT INTO ... VALUES  ON CONFLICT ...") — невалидный SQL,
	// поэтому SaveClicks обязан выйти до похода в пул вообще. poolMock без EXPECT() на
	// OpTimeout/Exec — если SaveClicks всё-таки попробует сходить в пул, gomock уронит тест.
	err := repository.SaveClicks(context.Background(), nil)

	assert.NoError(t, err)
}
