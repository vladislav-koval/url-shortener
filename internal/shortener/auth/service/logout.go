package service

import "context"

func (s *AuthService) Logout(ctx context.Context, rawToken string) {
	s.sessionManager.Delete(ctx, rawToken)
}
