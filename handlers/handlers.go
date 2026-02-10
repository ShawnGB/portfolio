package handlers

import (
	"log"
	"net/http"
	"sync"

	"mymodules/gofolio/i18n"
	"mymodules/gofolio/utils"
	"mymodules/gofolio/views/pages"
)

var (
	cachedImages []string
	imagesMutex  sync.RWMutex
)

func LoadImages() error {
	images, err := utils.GetArtsImages()
	if err != nil {
		return err
	}
	imagesMutex.Lock()
	cachedImages = images
	imagesMutex.Unlock()
	log.Printf("INFO: Loaded %d images for Arts page", len(images))
	return nil
}

func ArtsHandler(w http.ResponseWriter, r *http.Request) {
	pCtx := i18n.NewPageContext(r)

	imagesMutex.RLock()
	images := cachedImages
	imagesMutex.RUnlock()

	component := pages.Arts(images, pCtx)
	component.Render(r.Context(), w)
}
