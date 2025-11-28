package handlers

import (
	"log"
	"net/http"
	"sync"

	"mymodules/gofolio/i18n"
	"mymodules/gofolio/types"
	"mymodules/gofolio/utils"
	"mymodules/gofolio/views/pages"
)

var (
	cachedImages         []string
	imagesMutex          sync.RWMutex
	featuredProjects     []types.FeaturedProject
	featuredProjectMutex sync.RWMutex
	githubStats          *types.GitHubStats
	githubStatsMutex     sync.RWMutex
)

// LoadImages loads and caches image filenames at startup
func LoadImages() error {
	images, err := utils.GetImageFilenames()
	if err != nil {
		return err
	}
	imagesMutex.Lock()
	cachedImages = images
	imagesMutex.Unlock()
	log.Printf("INFO: Loaded %d images for Arts page", len(images))
	return nil
}

// ArtsHandler handles the Arts page with cached image list
func ArtsHandler(w http.ResponseWriter, r *http.Request) {
	pCtx := i18n.NewPageContext(r)

	imagesMutex.RLock()
	images := cachedImages
	imagesMutex.RUnlock()

	component := pages.Arts(images, pCtx)
	component.Render(r.Context(), w)
}

// LoadGitHubStats fetches and caches GitHub statistics at startup
func LoadGitHubStats() error {
	stats, err := utils.FetchGitHubStats("ShawnGB")
	if err != nil {
		log.Printf("WARN: Failed to fetch GitHub stats: %v", err)
		return err
	}

	githubStatsMutex.Lock()
	githubStats = stats
	githubStatsMutex.Unlock()

	return nil
}

// LoadFeaturedProjects initializes featured projects and fetches screenshots at startup
func LoadFeaturedProjects() error {
	projects := []types.FeaturedProject{
		{
			Name:         "jessebecker.com",
			URL:          "http://www.jessebecker.com/",
			Description:  "Minimal artist portfolio in pure HTML, CSS & JavaScript",
			Technologies: []string{"HTML", "CSS", "JavaScript"},
			ImagePath:    "/static/images/projects/jessebecker.png",
			Status:       "live",
		},
		{
			Name:         "arinashanzev.com",
			URL:          "https://arinashanzev.com",
			Description:  "WordPress portfolio for energy work and therapy practice",
			Technologies: []string{"WordPress", "PHP", "Custom Theme"},
			ImagePath:    "/static/images/projects/arinashanzev.png",
			Status:       "live",
		},
		{
			Name:         "berndwolf.net",
			URL:          "https://berndwolf.net/en/home",
			Description:  "Artist portfolio with custom WordPress extensions",
			Technologies: []string{"WordPress", "PHP", "JavaScript"},
			ImagePath:    "/static/images/projects/berndwolf.png",
			Status:       "live",
		},
		{
			Name:         "evolve-festival.com",
			URL:          "https://www.evolve-festival.com/",
			Description:  "Festival website with enhanced UX and ticket integration",
			Technologies: []string{"HTML", "CSS", "JavaScript", "Ticketing API"},
			ImagePath:    "/static/images/projects/evolve.png",
			Status:       "archived",
		},
	}

	// Fetch screenshots for all projects
	screenshotConfigs := []utils.ScreenshotConfig{
		{URL: "http://www.jessebecker.com/", Filename: "jessebecker.png"},
		{URL: "https://arinashanzev.com", Filename: "arinashanzev.png"},
		{URL: "https://berndwolf.net/en/home", Filename: "berndwolf.png"},
		{URL: "https://www.evolve-festival.com/", Filename: "evolve.png"},
	}

	utils.FetchAllScreenshots(screenshotConfigs)

	featuredProjectMutex.Lock()
	featuredProjects = projects
	featuredProjectMutex.Unlock()

	log.Printf("INFO: Loaded %d featured projects", len(projects))
	return nil
}

// GetFeaturedProjects returns the cached featured projects list
func GetFeaturedProjects() []types.FeaturedProject {
	featuredProjectMutex.RLock()
	defer featuredProjectMutex.RUnlock()
	return featuredProjects
}

// ProjectsHandler handles the Projects page with featured projects carousel
func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	pCtx := i18n.NewPageContext(r)

	featuredProjectMutex.RLock()
	projects := featuredProjects
	featuredProjectMutex.RUnlock()

	githubStatsMutex.RLock()
	stats := githubStats
	githubStatsMutex.RUnlock()

	component := pages.Projects(pCtx, projects, stats)
	component.Render(r.Context(), w)
}
