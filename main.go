package main

import (
	"log"
	"net/http"
	"os"

	"mymodules/gofolio/bundle"
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

	if err := bundle.LoadCSS("./static/styles.css"); err != nil {
		log.Fatalf("FATAL: bundling CSS: %v", err)
	}
	log.Printf("INFO: CSS bundled (version %s)", bundle.CSSVersion())
	mux.HandleFunc("/static/styles.css", bundle.ServeCSS)

	fs := http.FileServer(noDirListing{http.Dir("./static")})
	mux.Handle("/static/", http.StripPrefix("/static/", staticCache(fs)))

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

// staticCache lets browsers keep static files for a week. Files are not
// content-hashed, so this stays well short of "immutable".
func staticCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=604800")
		next.ServeHTTP(w, r)
	})
}

// noDirListing wraps a FileSystem so directories return 404 instead of an
// auto-generated file index.
type noDirListing struct{ fs http.FileSystem }

func (n noDirListing) Open(name string) (http.File, error) {
	f, err := n.fs.Open(name)
	if err != nil {
		return nil, err
	}
	if info, err := f.Stat(); err != nil || info.IsDir() {
		f.Close()
		return nil, os.ErrNotExist
	}
	return f, nil
}

func registerPageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			pCtx := i18n.NewPageContext(r)
			if err := pages.NotFound(pCtx).Render(r.Context(), w); err != nil {
				log.Printf("ERROR: rendering 404 page: %v", err)
			}
			return
		}
		handlers.PageHandler(pages.Home)(w, r)
	})
	mux.HandleFunc("/about", handlers.PageHandler(pages.About))
	mux.HandleFunc("/experience", handlers.PageHandler(pages.Experience))
	mux.HandleFunc("/projects", handlers.PageHandler(pages.Projects))
	mux.HandleFunc("/arts", handlers.ArtsHandler)
	mux.HandleFunc("/contact", handlers.ContactHandler)
	mux.HandleFunc("/impressum", handlers.PageHandler(pages.Impressum))
	mux.HandleFunc("/datenschutz", handlers.PageHandler(pages.Privacy))
}
