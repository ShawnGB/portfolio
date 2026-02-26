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

	handlers.InitCaptchaClient()
	handlers.InitEmailConfig()
	i18n.Init("i18n/locales")
	if err := handlers.LoadImages(); err != nil {
		log.Printf("WARN: Failed to load images: %v", err)
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/robots.txt")
	})

	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/sitemap.xml")
	})

	mux.HandleFunc("/.well-known/security.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/.well-known/security.txt")
	})

	registerPageRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("INFO: No PORT set, defaulting to :8080")
	}

	log.Printf("INFO: Starting server on :%s", port)

	handler := middleware.Recovery(middleware.Security(middleware.Logging(i18n.MiddlewareI18n(mux))))

	err = http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatalf("FATAL: Server error: %v", err)
	}
}

func registerPageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handlers.PageHandler(pages.Home))
	mux.HandleFunc("/about", handlers.PageHandler(pages.About))
	mux.HandleFunc("/experience", handlers.PageHandler(pages.Experience))
	mux.HandleFunc("/projects", handlers.PageHandler(pages.Projects))
	mux.HandleFunc("/arts", handlers.ArtsHandler)
	mux.HandleFunc("/contact", handlers.ContactHandler)
}
