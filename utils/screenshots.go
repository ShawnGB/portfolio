package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ScreenshotConfig holds configuration for screenshot fetching
type ScreenshotConfig struct {
	URL      string
	Filename string
}

// FetchScreenshot downloads a screenshot of a website using a screenshot API
// Saves to static/images/projects/{filename}
func FetchScreenshot(config ScreenshotConfig) error {
	// Create directory if it doesn't exist
	dir := "./static/images/projects"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	outputPath := filepath.Join(dir, config.Filename)

	// Check if screenshot already exists
	if _, err := os.Stat(outputPath); err == nil {
		log.Printf("Screenshot already exists: %s", outputPath)
		return nil
	}

	// Use Microlink API for screenshots (free tier, no auth required)
	// Alternative services: screenshotapi.net (requires key), htmlcsstoimage.com
	screenshotURL := fmt.Sprintf(
		"https://api.microlink.io/screenshot?url=%s&width=1200&height=800&type=png&viewport.width=1200&viewport.height=800",
		config.URL,
	)

	log.Printf("Fetching screenshot for %s...", config.URL)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(screenshotURL)
	if err != nil {
		return fmt.Errorf("failed to fetch screenshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("screenshot API returned status %d", resp.StatusCode)
	}

	// Create output file
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Write response to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("Screenshot saved: %s", outputPath)
	return nil
}

// FetchAllScreenshots fetches screenshots for all configured projects
func FetchAllScreenshots(configs []ScreenshotConfig) {
	for _, config := range configs {
		if err := FetchScreenshot(config); err != nil {
			log.Printf("Warning: Failed to fetch screenshot for %s: %v", config.URL, err)
		}
	}
}
