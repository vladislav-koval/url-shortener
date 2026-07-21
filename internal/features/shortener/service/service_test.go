package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/core/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/core/domain"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/service/mocks"
	"go.uber.org/mock/gomock"
)

func before(t *testing.T) (*mocks.MockRepository, *mocks.MockClickRecorder, *Service, func()) {
	ctrl := gomock.NewController(t)

	repository := mocks.NewMockRepository(ctrl)
	recorder := mocks.NewMockClickRecorder(ctrl)

	svc := NewService(repository, recorder)

	teardown := func() {
		ctrl.Finish()
	}

	return repository, recorder, svc, teardown

}

func TestCreatedShortLink_Successful(t *testing.T) {
	repository, _, svc, teardown := before(t)
	defer teardown()

	repository.EXPECT().CreateShortLink(gomock.Any(), gomock.Any(), "http://google.com").Return(domain.Link{
		ShortCode:   "shortLink",
		OriginalURL: "http://google.com",
	}, nil).Times(1)

	actual, err := svc.CreateShortLink(context.Background(), "http://google.com")

	assert.NoError(t, err)
	assert.Equal(t, domain.Link{
		ShortCode:   "shortLink",
		OriginalURL: "http://google.com",
	}, actual)
}

func TestCreateShortLink_RetryOnConflict(t *testing.T) {
	repository, _, svc, teardown := before(t)
	defer teardown()

	repository.EXPECT().
		CreateShortLink(gomock.Any(), gomock.Any(), "http://google.com").
		Return(domain.Link{}, apperrors.ErrConflict). // Первый вызов вернет конфликт
		Times(1)

	repository.EXPECT().
		CreateShortLink(gomock.Any(), gomock.Any(), "http://google.com").
		Return(domain.Link{
			ShortCode:   "successCode",
			OriginalURL: "http://google.com",
		}, nil).
		Times(1)

	actual, err := svc.CreateShortLink(context.Background(), "http://google.com")

	assert.NoError(t, err)
	assert.Equal(t, domain.Link{
		ShortCode:   "successCode",
		OriginalURL: "http://google.com",
	}, actual)
}

func TestCreateShortLink_Failed(t *testing.T) {
	repository, _, svc, teardown := before(t)
	defer teardown()

	repository.EXPECT().
		CreateShortLink(gomock.Any(), gomock.Any(), "http://google.com").
		Return(domain.Link{}, apperrors.ErrConflict). // Первый вызов вернет конфликт
		Times(5)

	_, err := svc.CreateShortLink(context.Background(), "http://google.com")

	assert.ErrorIs(t, err, apperrors.ErrConflict, "Должна вернуться ошибка ErrConflict")
}

func TestCreateShortLink_GeneratedCodeProperties(t *testing.T) {
	repo, _, svs, teardown := before(t)
	defer teardown()

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
