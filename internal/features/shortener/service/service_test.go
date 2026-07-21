package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/core/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/core/domain"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/service/mocks"
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
	tests := []struct {
		name        string
		originalURL string
		setupMock   func(repo *mocks.MockRepository)
		wantLink    domain.Link
		wantErr     error
	}{
		{
			name:        "successful",
			originalURL: "http://google.com",
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().
					CreateShortLink(gomock.Any(), gomock.Any(), "http://google.com").
					Return(domain.Link{ShortCode: "shortLink", OriginalURL: "http://google.com"}, nil).
					Times(1)
			},
			wantLink: domain.Link{ShortCode: "shortLink", OriginalURL: "http://google.com"},
		},
		{
			name:        "retry on conflict",
			originalURL: "http://google.com",
			setupMock: func(repo *mocks.MockRepository) {
				gomock.InOrder(
					repo.EXPECT(). // первый вызов возвращает конфликт
							CreateShortLink(gomock.Any(), gomock.Any(), "http://google.com").
							Return(domain.Link{}, apperrors.ErrConflict),
					repo.EXPECT(). // второй - успех
							CreateShortLink(gomock.Any(), gomock.Any(), "http://google.com").
							Return(domain.Link{ShortCode: "shortCode", OriginalURL: "http://google.com"}, nil),
				)
			},
			wantLink: domain.Link{ShortCode: "shortCode", OriginalURL: "http://google.com"},
		},
		{
			name:        "exhausted attempts",
			originalURL: "http://google.com",
			setupMock: func(repo *mocks.MockRepository) {
				repo.EXPECT().
					CreateShortLink(gomock.Any(), gomock.Any(), "http://google.com").
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository, _, svc := initTest(t)

			if tt.setupMock != nil {
				tt.setupMock(repository)
			}
			// иначе setupMock не задан - значит для этого кейса репозиторий
			// вообще не должен вызываться, и мок сам провалит тест, если это произойдёт

			actual, err := svc.CreateShortLink(context.Background(), tt.originalURL)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantLink, actual)
		})
	}
}

func TestCreateShortLink_GeneratedCodeProperties(t *testing.T) {
	repo, _, svs := initTest(t)

	repo.EXPECT().
		CreateShortLink(gomock.Any(), gomock.Any(), "http://google.com").
		DoAndReturn(func(ctx context.Context, code string, url string) (domain.Link, error) {

			assert.Len(t, code, 7)

			for _, char := range code {
				assert.Contains(t, shortCodeAlphabet, string(char), "Код содержит недопустимый символ!")
			}

			return domain.Link{ShortCode: code, OriginalURL: url}, nil
		}).
		Times(1)

	_, err := svs.CreateShortLink(context.Background(), "http://google.com")

	assert.NoError(t, err)
}

func TestResolveShortLink_Successful(t *testing.T) {
	repository, recorder, svc := initTest(t)

	repository.EXPECT().
		GetByShortCode(gomock.Any(), "shortCode").
		Return(domain.Link{
			ShortCode:   "shortCode",
			OriginalURL: "http://google.com",
		}, nil).
		Times(1)

	event := events.NewClickEvent("shortCode", new("127.0.0.1"))
	recorder.EXPECT().
		RecordClick(event).
		Return().
		Times(1)

	originalLink, err := svc.ResolveShortLink(context.Background(), "shortCode", event)

	assert.NoError(t, err)
	assert.Equal(t, "http://google.com", originalLink)
}

func TestResolveShortLink_ErrNotFound(t *testing.T) {
	repository, recorder, svc := initTest(t)

	repository.EXPECT().
		GetByShortCode(gomock.Any(), "shortCode").
		Return(domain.Link{}, apperrors.ErrNotFound).
		Times(1)

	recorder.EXPECT().
		RecordClick(gomock.Any()).
		Times(0)

	event := events.NewClickEvent("shortCode", new("127.0.0.1"))
	originalLink, err := svc.ResolveShortLink(context.Background(), "shortCode", event)

	assert.ErrorIs(t, err, apperrors.ErrNotFound)
	assert.Equal(t, "", originalLink)

}
