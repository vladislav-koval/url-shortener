package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/service/mocks"
)

func initTest(t *testing.T) (*mocks.MockIdentityProvider, *mocks.MockUserRepository, *mocks.MockSessionManager, *AuthService) {
	t.Helper()

	ctrl := gomock.NewController(t)

	identityProvider := mocks.NewMockIdentityProvider(ctrl)
	userRepository := mocks.NewMockUserRepository(ctrl)
	sessionManager := mocks.NewMockSessionManager(ctrl)

	svc := NewAuthService(identityProvider, userRepository, sessionManager)

	return identityProvider, userRepository, sessionManager, svc
}

func TestLoginWithGoogle(t *testing.T) {
	const (
		inputCode        = "auth-code"
		inputVerifier    = "verifier"
		inputSub         = "google-sub-1"
		inputEmail       = "user@example.com"
		inputName        = "User Name"
		wantSessionToken = "session-token"
	)

	savedUserID := uuid.New()

	testCases := []struct {
		name      string
		setupMock func(idp *mocks.MockIdentityProvider, repo *mocks.MockUserRepository, sm *mocks.MockSessionManager)
		check     func(t *testing.T, token string, err error)
	}{
		{
			name: "successful login",
			setupMock: func(idp *mocks.MockIdentityProvider, repo *mocks.MockUserRepository, sm *mocks.MockSessionManager) {
				idp.EXPECT().
					Exchange(gomock.Any(), inputCode, inputVerifier).
					Return(domain.GoogleIdentity{
						Sub:           inputSub,
						Email:         inputEmail,
						EmailVerified: true,
						Name:          inputName,
					}, nil).
					Times(1)

				repo.EXPECT().
					UpsertUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user domain.User) (domain.User, error) {
						assert.NotEqual(t, uuid.Nil, user.ID, "id must be generated before persisting")
						assert.Equal(t, inputSub, user.GoogleSub)
						assert.Equal(t, inputEmail, user.Email)
						assert.Equal(t, inputName, user.Name)

						return domain.User{ID: savedUserID, GoogleSub: user.GoogleSub, Email: user.Email}, nil
					}).
					Times(1)

				sm.EXPECT().
					Create(gomock.Any(), savedUserID).
					Return(wantSessionToken, nil).
					Times(1)
			},
			check: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, wantSessionToken, token)
			},
		},
		{
			name: "identity provider exchange fails",
			setupMock: func(idp *mocks.MockIdentityProvider, repo *mocks.MockUserRepository, sm *mocks.MockSessionManager) {
				idp.EXPECT().
					Exchange(gomock.Any(), inputCode, inputVerifier).
					Return(domain.GoogleIdentity{}, errors.New("google is down")).
					Times(1)

				repo.EXPECT().UpsertUser(gomock.Any(), gomock.Any()).Times(0)
				sm.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			check: func(t *testing.T, token string, err error) {
				assert.ErrorContains(t, err, "google is down")
				assert.Empty(t, token)
			},
		},
		{
			name: "incomplete identity - empty sub",
			setupMock: func(idp *mocks.MockIdentityProvider, repo *mocks.MockUserRepository, sm *mocks.MockSessionManager) {
				idp.EXPECT().
					Exchange(gomock.Any(), inputCode, inputVerifier).
					Return(domain.GoogleIdentity{Sub: "", Email: inputEmail, EmailVerified: true}, nil).
					Times(1)

				repo.EXPECT().UpsertUser(gomock.Any(), gomock.Any()).Times(0)
				sm.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			check: func(t *testing.T, token string, err error) {
				assert.ErrorIs(t, err, apperrors.ErrAuthorization)
				assert.Empty(t, token)
			},
		},
		{
			name: "incomplete identity - empty email",
			setupMock: func(idp *mocks.MockIdentityProvider, repo *mocks.MockUserRepository, sm *mocks.MockSessionManager) {
				idp.EXPECT().
					Exchange(gomock.Any(), inputCode, inputVerifier).
					Return(domain.GoogleIdentity{Sub: inputSub, Email: "", EmailVerified: true}, nil).
					Times(1)

				repo.EXPECT().UpsertUser(gomock.Any(), gomock.Any()).Times(0)
				sm.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			check: func(t *testing.T, token string, err error) {
				assert.ErrorIs(t, err, apperrors.ErrAuthorization)
				assert.Empty(t, token)
			},
		},
		{
			name: "email not verified",
			setupMock: func(idp *mocks.MockIdentityProvider, repo *mocks.MockUserRepository, sm *mocks.MockSessionManager) {
				idp.EXPECT().
					Exchange(gomock.Any(), inputCode, inputVerifier).
					Return(domain.GoogleIdentity{Sub: inputSub, Email: inputEmail, EmailVerified: false}, nil).
					Times(1)

				repo.EXPECT().UpsertUser(gomock.Any(), gomock.Any()).Times(0)
				sm.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			check: func(t *testing.T, token string, err error) {
				assert.ErrorIs(t, err, apperrors.ErrAuthorization)
				assert.Empty(t, token)
			},
		},
		{
			name: "upsert user fails",
			setupMock: func(idp *mocks.MockIdentityProvider, repo *mocks.MockUserRepository, sm *mocks.MockSessionManager) {
				idp.EXPECT().
					Exchange(gomock.Any(), inputCode, inputVerifier).
					Return(domain.GoogleIdentity{Sub: inputSub, Email: inputEmail, EmailVerified: true}, nil).
					Times(1)

				repo.EXPECT().
					UpsertUser(gomock.Any(), gomock.Any()).
					Return(domain.User{}, errors.New("db is down")).
					Times(1)

				sm.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
			},
			check: func(t *testing.T, token string, err error) {
				assert.ErrorContains(t, err, "db is down")
				assert.Empty(t, token)
			},
		},
		{
			name: "session creation fails",
			setupMock: func(idp *mocks.MockIdentityProvider, repo *mocks.MockUserRepository, sm *mocks.MockSessionManager) {
				idp.EXPECT().
					Exchange(gomock.Any(), inputCode, inputVerifier).
					Return(domain.GoogleIdentity{Sub: inputSub, Email: inputEmail, EmailVerified: true}, nil).
					Times(1)

				repo.EXPECT().
					UpsertUser(gomock.Any(), gomock.Any()).
					Return(domain.User{ID: savedUserID}, nil).
					Times(1)

				sm.EXPECT().
					Create(gomock.Any(), savedUserID).
					Return("", errors.New("redis is down")).
					Times(1)
			},
			check: func(t *testing.T, token string, err error) {
				assert.ErrorContains(t, err, "redis is down")
				assert.Empty(t, token)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			idp, repo, sm, svc := initTest(t)

			tt.setupMock(idp, repo, sm)

			token, err := svc.LoginWithGoogle(context.Background(), inputCode, inputVerifier)

			tt.check(t, token, err)
		})
	}
}

func TestLogout(t *testing.T) {
	_, _, sm, svc := initTest(t)

	const rawToken = "raw-token-value"

	sm.EXPECT().
		Delete(gomock.Any(), rawToken).
		Times(1)

	svc.Logout(context.Background(), rawToken)
}

func TestAuthCodeURL(t *testing.T) {
	idp, _, _, svc := initTest(t)

	const (
		state    = "state-value"
		verifier = "verifier-value"
		wantURL  = "https://accounts.google.com/o/oauth2/auth?state=state-value"
	)

	idp.EXPECT().
		AuthCodeURL(state, verifier).
		Return(wantURL).
		Times(1)

	got := svc.AuthCodeURL(state, verifier)

	assert.Equal(t, wantURL, got)
}
