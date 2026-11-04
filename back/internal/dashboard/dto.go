package dashboard

import "time"

type DashboardStatsResponse struct {
	TotalTests     int `json:"totalTests"`
	DraftTests     int `json:"draftTests"`
	PublishedTests int `json:"publishedTests"`
}

type RecentTestResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type DashboardResponse struct {
	Stats       DashboardStatsResponse `json:"stats"`
	RecentTests []RecentTestResponse   `json:"recentTests"`
}

func ToResponse(o DashboardOverview) DashboardResponse {
	recent := make([]RecentTestResponse, 0, len(o.RecentTests))
	for _, t := range o.RecentTests {
		recent = append(recent, RecentTestResponse{
			ID:        t.ID.String(),
			Title:     t.Title,
			Status:    string(t.Status),
			CreatedAt: t.CreatedAt,
		})
	}
	return DashboardResponse{
		Stats: DashboardStatsResponse{
			TotalTests:     o.Stats.TotalTests,
			DraftTests:     o.Stats.DraftTests,
			PublishedTests: o.Stats.PublishedTests,
		},
		RecentTests: recent,
	}
}