package handlers

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"mymodules/gofolio/i18n"
	"mymodules/gofolio/utils"
	"mymodules/gofolio/views/forms"
	"mymodules/gofolio/views/pages"

	"github.com/a-h/templ"
	"github.com/kataras/hcaptcha"
)

var (
	cachedImages  []string
	captchaClient *hcaptcha.Client
	emailConfig   utils.SendContactMailConfig
)

func InitCaptchaClient() {
	secret := os.Getenv("HCAPTCHA_SECRET_KEY")
	if secret == "" {
		log.Fatal("FATAL: HCAPTCHA_SECRET_KEY is not set. Cannot initialize hCaptcha client.")
	}
	if os.Getenv("HCAPTCHA_SITE_KEY") == "" {
		log.Println("WARN: HCAPTCHA_SITE_KEY is not set. Contact form captcha widget will not render.")
	}
	captchaClient = hcaptcha.New(secret)
	log.Println("INFO: hCaptcha client initialized successfully.")
}

func InitEmailConfig() {
	emailConfig = utils.SendContactMailConfig{
		ResendAPIKey:   os.Getenv("RESEND_API_KEY"),
		SenderEmail:    os.Getenv("SENDER_EMAIL"),
		RecipientEmail: os.Getenv("RECIPIENT_EMAIL"),
	}
	if emailConfig.ResendAPIKey == "" || emailConfig.SenderEmail == "" || emailConfig.RecipientEmail == "" {
		log.Println("WARN: Email environment variables not fully set. Contact form email will fail.")
	}
}

func LoadImages() error {
	images, err := utils.GetArtsImages()
	if err != nil {
		return err
	}
	cachedImages = images
	log.Printf("INFO: Loaded %d images for Arts page", len(images))
	return nil
}

func PageHandler(render func(i18n.PageContext) templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pCtx := i18n.NewPageContext(r)
		if err := render(pCtx).Render(r.Context(), w); err != nil {
			log.Printf("ERROR: rendering page component: %v", err)
		}
	}
}

func ArtsHandler(w http.ResponseWriter, r *http.Request) {
	pCtx := i18n.NewPageContext(r)

	images := cachedImages
	component := pages.Arts(images, pCtx)
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("ERROR: rendering Arts component: %v", err)
	}
}

func renderContact(w http.ResponseWriter, r *http.Request, component templ.Component, statusCode int) {
	w.WriteHeader(statusCode)
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("ERROR: rendering contact component: %v (status %d)", err, statusCode)
	}
}

func ContactHandler(w http.ResponseWriter, r *http.Request) {
	pCtx := i18n.NewPageContext(r)

	if r.Method == http.MethodGet {
		renderContact(w, r, forms.Contact(forms.ContactFormData{HCaptchaSiteKey: os.Getenv("HCAPTCHA_SITE_KEY")}, pCtx), http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var formData forms.ContactFormData
	formData.HCaptchaSiteKey = os.Getenv("HCAPTCHA_SITE_KEY")

	if !contactLimiter.allow(clientIP(r), time.Now()) {
		log.Println("WARN: contact form rate limit hit")
		formData.ErrorMessage = pCtx.T("contact.error.rateLimit")
		renderContact(w, r, forms.ContactFormPartial(formData, pCtx), http.StatusTooManyRequests)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxContactBody)
	if err := r.ParseForm(); err != nil {
		log.Printf("ERROR: Error parsing form: %v", err)
		formData.ErrorMessage = pCtx.T("contact.error.parse")
		renderContact(w, r, forms.ContactFormPartial(formData, pCtx), http.StatusInternalServerError)
		return
	}

	formData.Website = r.FormValue("website")
	formData.FirstName = strings.TrimSpace(r.FormValue("firstname"))
	formData.LastName = strings.TrimSpace(r.FormValue("lastname"))
	formData.Email = strings.TrimSpace(r.FormValue("email"))
	formData.Phone = strings.TrimSpace(r.FormValue("phone"))
	formData.Service = r.FormValue("service")
	formData.Message = strings.TrimSpace(r.FormValue("message"))

	if formData.Website != "" {
		log.Println("INFO: Honeypot field filled, contact form submission cancelled.")
		renderContact(w, r, forms.ContactFormPartial(forms.ContactFormData{
			SuccessMessage: pCtx.T("contact.success.message"),
		}, pCtx), http.StatusOK)
		return
	}

	if key := validateContact(formData.FirstName, formData.LastName, formData.Email, formData.Phone, formData.Service, formData.Message); key != "" {
		formData.ErrorMessage = pCtx.T(key)
		renderContact(w, r, forms.ContactFormPartial(formData, pCtx), http.StatusBadRequest)
		return
	}

	captchaResp := captchaClient.VerifyToken(r.FormValue("h-captcha-response"))
	if !captchaResp.Success {
		log.Printf("WARN: hCaptcha verification failed: %+v", captchaResp)
		formData.ErrorMessage = pCtx.T("contact.error.captcha")
		renderContact(w, r, forms.ContactFormPartial(formData, pCtx), http.StatusBadRequest)
		return
	}

	if emailConfig.ResendAPIKey == "" || emailConfig.SenderEmail == "" || emailConfig.RecipientEmail == "" {
		log.Println("ERROR: Email environment variables not (fully) set. Cannot send mail.")
		formData.ErrorMessage = pCtx.T("contact.error.config")
		renderContact(w, r, forms.ContactFormPartial(formData, pCtx), http.StatusInternalServerError)
		return
	}

	mailData := utils.ContactMailData{
		FirstName: formData.FirstName,
		LastName:  formData.LastName,
		Email:     formData.Email,
		Phone:     formData.Phone,
		Service:   formData.Service,
		Message:   formData.Message,
	}

	_, emailErr := utils.SendContactMail(emailConfig, mailData)
	if emailErr != nil {
		log.Printf("ERROR: Error from utils.SendContactMail: %v\n", emailErr)
		formData.ErrorMessage = pCtx.T("contact.error.send")
		renderContact(w, r, forms.ContactFormPartial(formData, pCtx), http.StatusInternalServerError)
		return
	}

	log.Println("INFO: Contact email sent.")
	renderContact(w, r, forms.ContactFormPartial(forms.ContactFormData{
		SuccessMessage: pCtx.T("contact.success.message"),
	}, pCtx), http.StatusOK)
}
