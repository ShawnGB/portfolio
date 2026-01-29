package middleware

import (
	"net/http"
)

// Security adds security headers to all HTTP responses
func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict browser features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy
		// Allows: self-hosted resources, inline scripts/styles (for templ), hCaptcha, Google Fonts, HTMX, Google Analytics
		// Note: hCaptcha requires *.hcaptcha.com wildcards as subdomains vary by region/time
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://hcaptcha.com https://*.hcaptcha.com https://unpkg.com https://www.googletagmanager.com https://www.google-analytics.com; " +
			"style-src 'self' 'unsafe-inline' https://hcaptcha.com https://*.hcaptcha.com https://fonts.googleapis.com; " +
			"frame-src https://hcaptcha.com https://*.hcaptcha.com; " +
			"connect-src 'self' https://hcaptcha.com https://*.hcaptcha.com https://*.google-analytics.com https://*.analytics.google.com https://*.googletagmanager.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https://*.google-analytics.com https://*.googletagmanager.com; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self';"
		w.Header().Set("Content-Security-Policy", csp)

		// HSTS - Force HTTPS for 1 year (only set if request is HTTPS)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}
