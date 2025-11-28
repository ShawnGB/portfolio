package types

import "time"

// FeaturedProject represents a featured project in the carousel
type FeaturedProject struct {
	Name         string
	URL          string
	Description  string
	Technologies []string
	ImagePath    string
	Status       string // "live", "archived", "wip"
}

// GitHubStats holds aggregated GitHub statistics
type GitHubStats struct {
	YearsOnGitHub   int
	TotalRepos      int
	PrimaryLanguage string
	Languages       []string
	LastUpdated     time.Time
}
