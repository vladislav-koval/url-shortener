package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mileusna/useragent"
	"github.com/stretchr/testify/assert"
	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/domain"
	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/service/mocks"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/events"
	"go.uber.org/mock/gomock"
)

func TestGetDeviceType(t *testing.T) {
	testCases := []struct {
		name string
		ua   useragent.UserAgent
		want string
	}{
		{
			name: "tablet",
			ua:   useragent.UserAgent{Tablet: true},
			want: "tablet",
		},
		{
			name: "mobile",
			ua:   useragent.UserAgent{Mobile: true},
			want: "mobile",
		},
		{
			name: "desktop",
			ua:   useragent.UserAgent{Desktop: true},
			want: "desktop",
		},
		{
			name: "none of the flags set",
			ua:   useragent.UserAgent{},
			want: "",
		},
		{
			name: "tablet takes priority over mobile",
			ua:   useragent.UserAgent{Tablet: true, Mobile: true},
			want: "tablet",
		},
		{
			name: "mobile takes priority over desktop",
			ua:   useragent.UserAgent{Mobile: true, Desktop: true},
			want: "mobile",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, getDeviceType(tc.ua))
		})
	}
}

func initTest(t *testing.T) (*mocks.MockRepository, *Service) {
	t.Helper()

	ctrl := gomock.NewController(t)
	repository := mocks.NewMockRepository(ctrl)
	svc := NewService(repository)

	return repository, svc
}

func TestCreateClicks(t *testing.T) {
	const (
		chromeWindowsUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
		safariIPhoneUA  = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
	)

	t.Run("maps events to clicks and saves them", func(t *testing.T) {
		repository, svc := initTest(t)

		inputEvents := []events.ClickEvent{
			{
				ID:          uuid.New(),
				ShortCode:   "shortCode-0",
				ClickedAt:   time.Now(),
				CountryCode: "US",
				City:        "New York",
				UserAgent:   chromeWindowsUA,
				Referer:     "referer-0.com",
			},
			{
				ID:          uuid.New(),
				ShortCode:   "shortCode-1",
				ClickedAt:   time.Now(),
				CountryCode: "DE",
				City:        "Berlin",
				UserAgent:   safariIPhoneUA,
				Referer:     "referer-1.com",
			},
			{
				ID:          uuid.New(),
				ShortCode:   "shortCode-2",
				ClickedAt:   time.Now(),
				CountryCode: "",
				City:        "",
				UserAgent:   "",
				Referer:     "",
			},
		}

		wantClicks := []domain.Click{
			{
				ID:        inputEvents[0].ID,
				ShortCode: "shortCode-0",
				ClickedAt: inputEvents[0].ClickedAt,

				Country: "US",
				City:    "New York",

				DeviceType: "desktop",
				OS:         "Windows",
				Browser:    "Chrome",

				Referer: "referer-0.com",
			},
			{
				ID:        inputEvents[1].ID,
				ShortCode: "shortCode-1",
				ClickedAt: inputEvents[1].ClickedAt,

				Country: "DE",
				City:    "Berlin",

				DeviceType: "mobile",
				OS:         "iOS",
				Browser:    "Safari",

				Referer: "referer-1.com",
			},
			{
				// Пустой UserAgent: useragent.Parse("") не даёт совпадений ни по одному
				// флагу устройства, ни по OS/браузеру — все три поля остаются пустыми,
				// без синтетического значения вроде "unknown".
				ID:        inputEvents[2].ID,
				ShortCode: "shortCode-2",
				ClickedAt: inputEvents[2].ClickedAt,

				Country: "",
				City:    "",

				DeviceType: "",
				OS:         "",
				Browser:    "",

				Referer: "",
			},
		}

		repository.EXPECT().
			SaveClicks(gomock.Any(), wantClicks).
			Return(nil).
			Times(1)

		err := svc.CreateClicks(context.Background(), inputEvents)

		assert.NoError(t, err)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		repository, svc := initTest(t)

		wantErr := errors.New("save failed")

		repository.EXPECT().
			SaveClicks(gomock.Any(), gomock.Any()).
			Return(wantErr).
			Times(1)

		err := svc.CreateClicks(context.Background(), []events.ClickEvent{
			{ID: uuid.New(), ShortCode: "shortCode-0", ClickedAt: time.Now()},
		})

		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("empty events still calls the repository", func(t *testing.T) {
		repository, svc := initTest(t)

		repository.EXPECT().
			SaveClicks(gomock.Any(), []domain.Click{}).
			Return(nil).
			Times(1)

		err := svc.CreateClicks(context.Background(), []events.ClickEvent{})

		assert.NoError(t, err)
	})
}
