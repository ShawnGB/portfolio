// Package bundle concatenates the CSS entry file and its @import chain into a
// single stylesheet at startup, so browsers make one request instead of a
// waterfall of 15. The content hash doubles as a cache-busting version.
package bundle

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var importRe = regexp.MustCompile(`^@import\s+(?:url\()?["']([^"']+)["']\)?\s*;`)

var (
	css     []byte
	version string
)

// LoadCSS reads entry (e.g. ./static/styles.css), inlines its local @imports
// recursively, and keeps the result in memory.
func LoadCSS(entry string) error {
	var buf bytes.Buffer
	if err := inline(&buf, entry, map[string]bool{}); err != nil {
		return err
	}
	css = buf.Bytes()
	sum := sha256.Sum256(css)
	version = hex.EncodeToString(sum[:])[:10]
	return nil
}

func inline(buf *bytes.Buffer, path string, seen map[string]bool) error {
	path = filepath.Clean(path)
	if seen[path] {
		return fmt.Errorf("css import cycle at %s", path)
	}
	seen[path] = true
	defer delete(seen, path)

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if m := importRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil && !strings.Contains(m[1], "://") {
			fmt.Fprintf(buf, "/* %s */\n", m[1])
			if err := inline(buf, filepath.Join(filepath.Dir(path), m[1]), seen); err != nil {
				return err
			}
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	return sc.Err()
}

// CSSVersion is the short content hash of the bundled stylesheet.
func CSSVersion() string { return version }

// ServeCSS serves the bundled stylesheet. Requests carrying the current
// version are cached for a year; anything else must revalidate.
func ServeCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	if r.URL.Query().Get("v") == version {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	w.Header().Set("ETag", `"`+version+`"`)
	if r.Header.Get("If-None-Match") == `"`+version+`"` {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Write(css)
}
