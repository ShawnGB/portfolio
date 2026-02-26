package utils

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// walkImages walks dir and returns image paths with trimPrefix removed.
// filepath.Join (used internally by WalkDir) calls filepath.Clean, which
// strips any leading "./", so we normalise both forms here.
func walkImages(dir, trimPrefix string) ([]string, error) {
	var images []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && isImage(path) {
			clean := strings.TrimPrefix(path, "./")
			images = append(images, strings.TrimPrefix(clean, trimPrefix))
		}
		return nil
	})
	return images, err
}

func GetArtsImages() ([]string, error) { return walkImages("./static/images/arts", "static/") }

func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}
