package session

import "context"

func (s *Service) Delete(
	ctx context.Context,
	rawToken string,
) {
	tokenHash := hashSessionToken(rawToken)

	s.repository.Delete(ctx, tokenHash)
}
