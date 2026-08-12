package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/geo"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/service/mocks"
	"go.uber.org/mock/gomock"
)

func initTest(t *testing.T) (*mocks.MockRepository, *mocks.MockClickRecorder, *Service) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repository := mocks.NewMockRepository(ctrl)
	recorder := mocks.NewMockClickRecorder(ctrl)
	svc := NewService(repository, recorder)

	return repository, recorder, svc
}

func TestCreateShortLink(t *testing.T) {
	const (
		inputShortCode   = "shortCode"
		inputOriginalURL = "http://google.com"
	)

	testCases := []struct {
		name        string
		originalURL string
		setupMock   func(repo *mocks.MockRepository, originalURL string)
		wantLink    domain.Link
		wantErr     error
	}{
		{
			name:        "successful",
			originalURL: inputOriginalURL,
			setupMock: func(repo *mocks.MockRepository, originalURL string) {
				repo.EXPECT().
					CreateShortLink(gomock.Any(), gomock.Any(), originalURL, gomock.Any()).
					Return(domain.Link{ShortCode: inputShortCode, OriginalURL: originalURL}, nil).
					Times(1)
			},
			wantLink: domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL},
		},
		{
			name:        "retry on conflict",
			originalURL: inputOriginalURL,
			setupMock: func(repo *mocks.MockRepository, originalURL string) {
				gomock.InOrder(
					repo.EXPECT().
						CreateShortLink(gomock.Any(), gomock.Any(), originalURL, gomock.Any()).
						Return(domain.Link{}, apperrors.ErrConflict),
					repo.EXPECT().
						CreateShortLink(gomock.Any(), gomock.Any(), originalURL, gomock.Any()).
						Return(domain.Link{ShortCode: inputShortCode, OriginalURL: originalURL}, nil),
				)
			},
			wantLink: domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL},
		},
		{
			name:        "exhausted attempts",
			originalURL: inputOriginalURL,
			setupMock: func(repo *mocks.MockRepository, originalURL string) {
				repo.EXPECT().
					CreateShortLink(gomock.Any(), gomock.Any(), originalURL, gomock.Any()).
					Return(domain.Link{}, apperrors.ErrConflict).
					Times(maxCreateLinkAttempts)
			},
			wantErr: apperrors.ErrConflict,
		},
		{
			name:        "invalid url - empty string fails to parse",
			originalURL: "",
			wantErr:     apperrors.ErrInvalidArgument,
		},
		{
			name:        "invalid url - unsupported scheme",
			originalURL: "ftp://example.com",
			wantErr:     apperrors.ErrInvalidArgument,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			repository, _, svc := initTest(t)

			if tt.setupMock != nil {
				tt.setupMock(repository, tt.originalURL)
			}

			actual, err := svc.CreateShortLink(context.Background(), tt.originalURL, nil)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantLink, actual)
		})
	}

	t.Run("generated code properties", func(t *testing.T) {
		repo, _, svc := initTest(t)

		repo.EXPECT().
			CreateShortLink(gomock.Any(), gomock.Any(), inputOriginalURL, gomock.Any()).
			DoAndReturn(func(ctx context.Context, code string, url string, userID *uuid.UUID) (domain.Link, error) {
				assert.Len(t, code, 7) // Хардкод длины защищает от случайного изменения константы

				for _, char := range code {
					assert.Contains(t, shortCodeAlphabet, string(char), "Код содержит недопустимый символ!")
				}

				return domain.Link{ShortCode: code, OriginalURL: url}, nil
			}).
			Times(1)

		_, err := svc.CreateShortLink(context.Background(), inputOriginalURL, nil)
		assert.NoError(t, err)
	})

	t.Run("passes userID through to repository", func(t *testing.T) {
		repo, _, svc := initTest(t)

		userID := uuid.New()

		repo.EXPECT().
			CreateShortLink(gomock.Any(), gomock.Any(), inputOriginalURL, &userID).
			Return(domain.Link{ShortCode: inputShortCode, OriginalURL: inputOriginalURL, UserID: &userID}, nil).
			Times(1)

		actual, err := svc.CreateShortLink(context.Background(), inputOriginalURL, &userID)

		assert.NoError(t, err)
		assert.Equal(t, &userID, actual.UserID)
	})
}

func TestResolveShortLink(t *testing.T) {
	const (
		inputShortCode   = "shortCode"
		inputOriginalURL = "http://google.com"
		inputCountry     = "US"
		inputCity        = "New York"
	)

	t.Run("successful resolution and click recording", func(t *testing.T) {
		repo, recorder, svc := initTest(t)

		repo.EXPECT().
			GetByShortCode(gomock.Any(), inputShortCode).
			Return(domain.Link{
				ShortCode:   inputShortCode,
				OriginalURL: inputOriginalURL,
			}, nil).
			Times(1)

		event := events.NewClickEvent(inputShortCode, geo.Geo{Country: inputCountry, City: inputCity})

		recorder.EXPECT().
			RecordClick(event).
			Times(1)

		originalLink, err := svc.ResolveShortLink(context.Background(), inputShortCode, event)

		assert.NoError(t, err)
		assert.Equal(t, inputOriginalURL, originalLink)
	})

	t.Run("short link not found", func(t *testing.T) {
		repo, recorder, svc := initTest(t)

		repo.EXPECT().
			GetByShortCode(gomock.Any(), inputShortCode).
			Return(domain.Link{}, apperrors.ErrNotFound).
			Times(1)

		recorder.EXPECT().
			RecordClick(gomock.Any()).
			Times(0)

		event := events.NewClickEvent(inputShortCode, geo.Geo{Country: inputCountry, City: inputCity})
		originalLink, err := svc.ResolveShortLink(context.Background(), inputShortCode, event)

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Equal(t, "", originalLink)
	})
}
