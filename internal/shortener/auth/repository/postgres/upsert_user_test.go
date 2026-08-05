package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool/mocks"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

func initTest(t *testing.T) (*Repository, *mocks.MockPool, *mocks.MockRow) {
	t.Helper()

	ctrl := gomock.NewController(t)

	poolMock := mocks.NewMockPool(ctrl)

	poolMock.EXPECT().
		OpTimeout().
		Return(5 * time.Second).
		Times(1)

	rowMock := mocks.NewMockRow(ctrl)

	repository := NewUserRepository(poolMock)

	return repository, poolMock, rowMock
}

func TestUpsertUser(t *testing.T) {
	inputUser := domain.User{
		ID:            uuid.New(),
		GoogleSub:     "google-sub-1",
		Email:         "user@example.com",
		EmailVerified: true,
		Name:          "User Name",
	}

	now := time.Now()

	testCases := []struct {
		name     string
		scanMock func(dest ...interface{}) error
		check    func(t *testing.T, user domain.User, err error)
	}{
		{
			name: "success upsert",
			scanMock: func(dest ...interface{}) error {
				*dest[0].(*uuid.UUID) = inputUser.ID
				*dest[1].(*string) = inputUser.GoogleSub
				*dest[2].(*string) = inputUser.Email
				*dest[3].(*bool) = inputUser.EmailVerified
				*dest[4].(*string) = inputUser.Name
				*dest[5].(*time.Time) = now
				*dest[6].(*time.Time) = now

				return nil
			},
			check: func(t *testing.T, user domain.User, err error) {
				assert.NoError(t, err)
				assert.Equal(t, inputUser.ID, user.ID)
				assert.Equal(t, inputUser.GoogleSub, user.GoogleSub)
				assert.Equal(t, inputUser.Email, user.Email)
				assert.Equal(t, inputUser.EmailVerified, user.EmailVerified)
				assert.Equal(t, inputUser.Name, user.Name)
				assert.Equal(t, now, user.CreatedAt)
				assert.Equal(t, now, user.UpdatedAt)
			},
		},
		{
			name: "scan error",
			scanMock: func(dest ...interface{}) error {
				return errors.New("something went wrong")
			},
			check: func(t *testing.T, _ domain.User, err error) {
				assert.EqualError(t, err, "upsert user: something went wrong")
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			repo, poolMock, rowMock := initTest(t)

			rowMock.EXPECT().
				Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(tt.scanMock).
				Times(1)

			poolMock.EXPECT().
				QueryRow(
					gomock.Any(),
					gomock.Any(),
					inputUser.ID,
					inputUser.GoogleSub,
					inputUser.Email,
					inputUser.EmailVerified,
					inputUser.Name,
				).
				Return(rowMock).
				Times(1)

			user, err := repo.UpsertUser(context.Background(), inputUser)

			tt.check(t, user, err)
		})
	}
}
