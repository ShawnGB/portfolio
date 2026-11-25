package handlers

import (
	"net"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	maxContactBody = 64 << 10 // 64 KB is far more than any honest submission
	contactLimit   = 5        // submissions per IP ...
	contactWindow  = 10 * time.Minute
	maxNameLen     = 100
	maxEmailLen    = 254
	maxPhoneLen    = 40
	maxMessageLen  = 5000
)

var allowedServices = map[string]bool{
	"": true, "entwicklung": true, "beratung": true, "product_management": true, "andere": true,
}

// rateLimiter is a small in-memory sliding-window limiter. It resets on
// restart, which is fine for a single-instance portfolio site.
type rateLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

var contactLimiter = &rateLimiter{hits: map[string][]time.Time{}}

func (l *rateLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-contactWindow)
	recent := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= contactLimit {
		l.hits[key] = recent
		return false
	}
	l.hits[key] = append(recent, now)

	// Opportunistic cleanup so the map can't grow without bound.
	if len(l.hits) > 1000 {
		for k, ts := range l.hits {
			if len(ts) == 0 || ts[len(ts)-1].Before(cutoff) {
				delete(l.hits, k)
			}
		}
	}
	return true
}

// clientIP uses the first X-Forwarded-For hop (set by Render's proxy) and
// falls back to the connection address.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func tooLong(s string, max int) bool { return utf8.RuneCountInString(s) > max }

// validateContact returns the locale key of the first problem, or "".
func validateContact(firstName, lastName, email, phone, service, message string) string {
	if firstName == "" || lastName == "" || email == "" || message == "" {
		return "contact.error.required"
	}
	if tooLong(firstName, maxNameLen) || tooLong(lastName, maxNameLen) || tooLong(email, maxEmailLen) ||
		tooLong(phone, maxPhoneLen) || tooLong(message, maxMessageLen) {
		return "contact.error.tooLong"
	}
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email || addr.Name != "" {
		return "contact.error.email"
	}
	if !allowedServices[service] {
		return "contact.error.required"
	}
	return ""
}
