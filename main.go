package main

import (
	"log"
	"net/http"
	"os"

	"mymodules/gofolio/handlers"
	"mymodules/gofolio/i18n"
	"mymodules/gofolio/middleware"
	"mymodules/gofolio/views/pages"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("INFO: .env not found")
	}

	// Initialize dependencies
	handlers.InitCaptchaClient()
	i18n.Init("i18n/locales")
	if err := handlers.LoadImages(); err != nil {
		log.Printf("WARN: Failed to load images: %v", err)
	}
	if err := handlers.LoadFeaturedProjects(); err != nil {
		log.Printf("WARN: Failed to load featured projects: %v", err)
	}
	if err := handlers.LoadGitHubStats(); err != nil {
		log.Printf("WARN: Failed to load GitHub stats: %v", err)
	}

	mux := http.NewServeMux()

	// Static File Server
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// routes fuer sitemap und robots
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/robots.txt")
	})

	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/sitemap.xml")
	})

	// Page routes using handler factory pattern
	registerPageRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		log.Printf("INFO: No PORT environment variable detected, defaulting to :%s", port)
	}

	log.Printf("INFO: Starting server on :%s", port)

	// Apply middleware chain: Recovery -> Logging -> i18n -> Routes
	handler := middleware.Recovery(middleware.Logging(i18n.MiddlewareI18n(mux)))

	err = http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatalf("FATAL: Server error: %v", err)
	}
}

func registerPageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handlers.NewPageHandler(pages.Home))
	mux.HandleFunc("/about", handlers.NewPageHandler(pages.About))
	mux.HandleFunc("/experience", handlers.NewPageHandler(pages.Experience))
	mux.HandleFunc("/projects", handlers.ProjectsHandler)
	mux.HandleFunc("/arts", handlers.ArtsHandler)
	mux.HandleFunc("/contact", handlers.ContactHandler)
}
