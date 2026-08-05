package service

func (s *AuthService) AuthCodeURL(state, verifier string) string {
	return s.identityProvider.AuthCodeURL(state, verifier)
}
