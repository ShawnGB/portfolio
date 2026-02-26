package i18n

import (
	"context"
	"html"
	"io"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

type PageContext struct {
	Localizer     *goi18n.Localizer
	ActiveLang    string
	BasePath      string
	OriginalQuery string
}

func (pc PageContext) T(messageID string) string {
	if pc.Localizer == nil {
		return messageID
	}
	return pc.Localizer.MustLocalize(&goi18n.LocalizeConfig{MessageID: messageID})
}

func (pc PageContext) TRich(messageID string) templ.Component {
	raw := pc.T(messageID)
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		parts := strings.Split(raw, "**")
		for i, part := range parts {
			var s string
			if i%2 == 1 {
				s = "<strong>" + html.EscapeString(part) + "</strong>"
			} else {
				s = html.EscapeString(part)
			}
			if _, err := io.WriteString(w, s); err != nil {
				return err
			}
		}
		return nil
	})
}

func buildLangURL(lang, basePath, query string) templ.SafeURL {
	path := "/" + lang
	if basePath != "/" {
		path += basePath
	} else {
		path += "/"
	}
	if query != "" {
		path += "?" + query
	}
	return templ.URL(path)
}

func (pc PageContext) LanguageSwitchURL(targetLangCode string) templ.SafeURL {
	return buildLangURL(targetLangCode, pc.BasePath, pc.OriginalQuery)
}

func (pc PageContext) CurrentLangLink(pathSegment string) templ.SafeURL {
	if pathSegment == "" {
		pathSegment = "/"
	} else if pathSegment[0] != '/' {
		pathSegment = "/" + pathSegment
	}
	return buildLangURL(pc.ActiveLang, pathSegment, "")
}

func NewPageContext(r *http.Request) PageContext {
	localizer := getLocalizer(r.Context())
	activeLang := getActiveLang(r.Context())
	basePath := getBasePath(r.Context())
	originalQuery := r.URL.RawQuery

	return PageContext{
		Localizer:     localizer,
		ActiveLang:    activeLang,
		BasePath:      basePath,
		OriginalQuery: originalQuery,
	}
}
