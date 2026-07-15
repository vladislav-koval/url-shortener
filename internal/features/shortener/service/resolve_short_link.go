package service

import (
	"context"
	"fmt"
)

func (s *Service) ResolveShortLink(ctx context.Context, code string) (string, error) {
	link, err := s.shortenerRepository.GetByShortCode(ctx, code)

	if err != nil {
		return "", fmt.Errorf("get link from repo: %w", err)
	}

	return link.OriginalURL, nil
}
