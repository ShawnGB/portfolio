package middleware

import (
	"net/http"
)

func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")

		w.Header().Set("X-Content-Type-Options", "nosniff")

		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

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

		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}
