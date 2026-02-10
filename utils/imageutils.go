package utils

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func GetImageFilenames() ([]string, error) {
	var images []string
	err := filepath.WalkDir("./static/images", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && isImage(path) {
			rel := strings.TrimPrefix(path, "./static/")
			images = append(images, rel)
		}
		return nil
	})

	return images, err
}

func GetArtsImages() ([]string, error) {
	var images []string
	err := filepath.WalkDir("./static/images/arts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && isImage(path) {
			rel := strings.TrimPrefix(path, "./static/")
			rel = strings.TrimPrefix(rel, "static/")
			images = append(images, rel)
		}
		return nil
	})

	return images, err
}

func GetProjectImages() ([]string, error) {
	var images []string
	err := filepath.WalkDir("./static/images/projects", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && isImage(path) {
			rel := strings.TrimPrefix(path, "./static/")
			images = append(images, rel)
		}
		return nil
	})

	return images, err
}

func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}
