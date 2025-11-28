package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mymodules/gofolio/types"
	"net/http"
	"os"
	"sort"
	"time"
)

// GitHubUser represents GitHub user API response
type GitHubUser struct {
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
	PublicRepos int     `json:"public_repos"`
}

// GitHubRepo represents a repository from GitHub API
type GitHubRepo struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Fork     bool   `json:"fork"`
}

var (
	cachedGitHubStats *types.GitHubStats
	statsCache        time.Time
	cacheDuration     = 24 * time.Hour
)

// FetchGitHubStats fetches GitHub statistics for a user
// Uses GITHUB_TOKEN env var if available for higher rate limits
func FetchGitHubStats(username string) (*types.GitHubStats, error) {
	// Return cached stats if still valid
	if cachedGitHubStats != nil && time.Since(statsCache) < cacheDuration {
		log.Println("INFO: Using cached GitHub stats")
		return cachedGitHubStats, nil
	}

	log.Println("INFO: Fetching fresh GitHub stats...")

	client := &http.Client{Timeout: 15 * time.Second}

	// Fetch user info
	userURL := fmt.Sprintf("https://api.github.com/users/%s", username)
	userReq, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}

	// Add auth token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		userReq.Header.Set("Authorization", "Bearer "+token)
	}
	userReq.Header.Set("Accept", "application/vnd.github.v3+json")

	userResp, err := client.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user data: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userResp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", userResp.StatusCode, string(body))
	}

	var user GitHubUser
	if err := json.NewDecoder(userResp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user data: %w", err)
	}

	// Fetch repositories
	reposURL := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&type=owner", username)
	reposReq, err := http.NewRequest("GET", reposURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create repos request: %w", err)
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		reposReq.Header.Set("Authorization", "Bearer "+token)
	}
	reposReq.Header.Set("Accept", "application/vnd.github.v3+json")

	reposResp, err := client.Do(reposReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repos: %w", err)
	}
	defer reposResp.Body.Close()

	if reposResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(reposResp.Body)
		return nil, fmt.Errorf("GitHub repos API returned status %d: %s", reposResp.StatusCode, string(body))
	}

	var repos []GitHubRepo
	if err := json.NewDecoder(reposResp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode repos: %w", err)
	}

	// Calculate stats
	stats := &types.GitHubStats{
		TotalRepos:  user.PublicRepos,
		LastUpdated: time.Now(),
	}

	// Calculate years on GitHub
	yearsOnGitHub := time.Since(user.CreatedAt).Hours() / 24 / 365
	stats.YearsOnGitHub = int(yearsOnGitHub)

	// Count languages (excluding forks and null languages)
	languageCounts := make(map[string]int)
	for _, repo := range repos {
		if !repo.Fork && repo.Language != "" {
			languageCounts[repo.Language]++
		}
	}

	// Get unique languages sorted by frequency
	type langCount struct {
		lang  string
		count int
	}
	var langCounts []langCount
	for lang, count := range languageCounts {
		langCounts = append(langCounts, langCount{lang, count})
	}
	sort.Slice(langCounts, func(i, j int) bool {
		return langCounts[i].count > langCounts[j].count
	})

	// Extract top languages
	stats.Languages = []string{}
	for i, lc := range langCounts {
		stats.Languages = append(stats.Languages, lc.lang)
		if i == 0 {
			stats.PrimaryLanguage = lc.lang
		}
	}

	// Cache the results
	cachedGitHubStats = stats
	statsCache = time.Now()

	log.Printf("INFO: GitHub stats fetched - %d years, %d repos, primary language: %s",
		stats.YearsOnGitHub, stats.TotalRepos, stats.PrimaryLanguage)

	return stats, nil
}

// GetCachedGitHubStats returns cached stats or nil if not available
func GetCachedGitHubStats() *types.GitHubStats {
	if cachedGitHubStats != nil && time.Since(statsCache) < cacheDuration {
		return cachedGitHubStats
	}
	return nil
}
