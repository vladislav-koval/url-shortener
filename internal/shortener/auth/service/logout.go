package service

import "context"

func (s *Service) Logout(ctx context.Context, rawToken string) {
	s.sessionManager.Delete(ctx, rawToken)
}
