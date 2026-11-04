package dashboard

import (
	"context"

	"github.com/google/uuid"
)

const recentTestsLimit = 5

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetOverview(ctx context.Context, ownerID uuid.UUID) (DashboardOverview, error) {
	stats, err := s.repo.GetStats(ctx, ownerID)
	if err != nil {
		return DashboardOverview{}, err
	}

	recent, err := s.repo.GetRecentTests(ctx, ownerID, recentTestsLimit)
	if err != nil {
		return DashboardOverview{}, err
	}
	if recent == nil {
		recent = []RecentTest{}
	}

	return DashboardOverview{Stats: stats, RecentTests: recent}, nil
}