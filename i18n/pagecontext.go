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
	L             *goi18n.Localizer
	ActiveLang    string
	BasePath      string
	OriginalQuery string
}

func (pc PageContext) T(messageID string) string {
	if pc.L == nil {
		return messageID
	}
	return pc.L.MustLocalize(&goi18n.LocalizeConfig{MessageID: messageID})
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

func (pc PageContext) LanguageSwitchURL(targetLangCode string) templ.SafeURL {
	path := "/" + targetLangCode
	if pc.BasePath != "/" {
		path += pc.BasePath
	} else {
		path += "/"
	}

	if pc.OriginalQuery != "" {
		path = path + "?" + pc.OriginalQuery
	}
	return templ.URL(path)
}

func (pc PageContext) CurrentLangLink(pathSegment string) templ.SafeURL {
	if pathSegment == "" {
		pathSegment = "/"
	} else if pathSegment[0] != '/' {
		pathSegment = "/" + pathSegment
	}

	path := "/" + pc.ActiveLang
	if pathSegment != "/" {
		path += pathSegment
	} else {
		path += "/"
	}
	return templ.URL(path)
}

func NewPageContext(r *http.Request) PageContext {
	localizer := GetLocalizer(r.Context())
	activeLang := GetActiveLang(r.Context())
	basePath := GetBasePath(r.Context())
	originalQuery := r.URL.RawQuery

	return PageContext{
		L:             localizer,
		ActiveLang:    activeLang,
		BasePath:      basePath,
		OriginalQuery: originalQuery,
	}
}
