package handlers

import (
	"net/http"

	"mymodules/gofolio/i18n"

	"github.com/a-h/templ"
)

type PageRenderer func(i18n.PageContext) templ.Component

func NewPageHandler(renderer PageRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pCtx := i18n.NewPageContext(r)
		component := renderer(pCtx)
		component.Render(r.Context(), w)
	}
}
