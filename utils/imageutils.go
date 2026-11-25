package utils

import (
	"encoding/binary"
	"errors"
	"io"
	"io/fs"
	"os"
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

// Image is a gallery image with its pixel size, so templates can set
// width/height and the browser can reserve space before it loads.
type Image struct {
	Path          string
	Width, Height int
}

// GetArtsImagesWithSize returns the Arts images with their dimensions.
// Images whose size can't be read are kept with 0×0 (rendered without
// width/height) rather than dropped.
func GetArtsImagesWithSize() ([]Image, error) {
	paths, err := GetArtsImages()
	if err != nil {
		return nil, err
	}
	out := make([]Image, 0, len(paths))
	for _, p := range paths {
		w, h, _ := webpSize(filepath.Join("static", p))
		out = append(out, Image{Path: p, Width: w, Height: h})
	}
	return out, nil
}

// webpSize reads the canvas size from a WebP file header (VP8, VP8L or
// VP8X), without decoding the image or pulling in an image library.
func webpSize(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	var b [30]byte
	if _, err := io.ReadFull(f, b[:]); err != nil {
		return 0, 0, err
	}
	if string(b[0:4]) != "RIFF" || string(b[8:12]) != "WEBP" {
		return 0, 0, errors.New("not a webp file")
	}
	switch string(b[12:16]) {
	case "VP8 ": // lossy: 14-bit width/height after the frame start code
		if b[23] != 0x9d || b[24] != 0x01 || b[25] != 0x2a {
			return 0, 0, errors.New("bad VP8 start code")
		}
		w := int(binary.LittleEndian.Uint16(b[26:28]) & 0x3fff)
		h := int(binary.LittleEndian.Uint16(b[28:30]) & 0x3fff)
		return w, h, nil
	case "VP8L": // lossless: 14-bit (size-1) fields after the 0x2f signature
		if b[20] != 0x2f {
			return 0, 0, errors.New("bad VP8L signature")
		}
		bits := binary.LittleEndian.Uint32(b[21:25])
		return int(bits&0x3fff) + 1, int((bits>>14)&0x3fff) + 1, nil
	case "VP8X": // extended: 24-bit (size-1) canvas fields
		w := int(b[24]) | int(b[25])<<8 | int(b[26])<<16
		h := int(b[27]) | int(b[28])<<8 | int(b[29])<<16
		return w + 1, h + 1, nil
	}
	return 0, 0, errors.New("unknown webp chunk")
}
