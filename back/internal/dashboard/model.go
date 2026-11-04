package dashboard

import (
	"time"

	"github.com/google/uuid"
)

type TestStatus string

const (
	StatusDraft     TestStatus = "draft"
	StatusPublished TestStatus = "published"
)

type DashboardStats struct {
	TotalTests     int
	DraftTests     int
	PublishedTests int
}

type RecentTest struct {
	ID        uuid.UUID
	Title     string
	Status    TestStatus
	CreatedAt time.Time
}

type DashboardOverview struct {
	Stats       DashboardStats
	RecentTests []RecentTest
}