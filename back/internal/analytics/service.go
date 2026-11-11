package analytics

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repositoty
}

func NewService(repo *Repositoty) *Service {
	return &Service{repo: repo}
}

// последние попытки по всем тестам преподавателя
func (s *Service) Recent(ctx context.Context, ownerID uuid.UUID, limit int) ([]RecentAttemptDTO, error){
	attempts, err := s.repo.Recent(ctx, ownerID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]RecentAttemptDTO, len(attempts))
	for i, att := range attempts{
		out[i] = att.ToDTO()
	}

	return out, nil
}