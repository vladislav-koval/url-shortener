package service

func (s *Service) AuthCodeURL(state, verifier string) string {
	return s.identityProvider.AuthCodeURL(state, verifier)
}
