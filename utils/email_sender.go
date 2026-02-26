package utils

import (
	"fmt"
	"html"
	"log"

	"github.com/resend/resend-go/v2"
)

type SendContactMailConfig struct {
	ResendAPIKey   string
	SenderEmail    string
	RecipientEmail string
}

type ContactMailData struct {
	FirstName, LastName, Email, Phone, Service, Message string
}

func SendContactMail(config SendContactMailConfig, data ContactMailData) (*resend.SendEmailResponse, error) {
	if config.ResendAPIKey == "" {
		return nil, fmt.Errorf("resend API Key missing")
	}
	if config.SenderEmail == "" {
		return nil, fmt.Errorf("sender email not configured")
	}
	if config.RecipientEmail == "" {
		return nil, fmt.Errorf("recipient not configured")
	}

	client := resend.NewClient(config.ResendAPIKey)

	htmlBody := fmt.Sprintf(`
		<h1>Neue Kontaktanfrage</h1>
		<p><strong>Vorname:</strong> %s</p>
		<p><strong>Nachname:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>
		<p><strong>Telefon:</strong> %s</p>
		<p><strong>Serviceleistung:</strong> %s</p>
		<p><strong>Nachricht:</strong></p>
		<p>%s</p>
	`,
		html.EscapeString(data.FirstName),
		html.EscapeString(data.LastName),
		html.EscapeString(data.Email),
		html.EscapeString(data.Phone),
		html.EscapeString(data.Service),
		html.EscapeString(data.Message),
	)

	params := &resend.SendEmailRequest{
		From:    config.SenderEmail,
		To:      []string{config.RecipientEmail},
		Subject: fmt.Sprintf("Neue Kontaktanfrage von %s %s", html.EscapeString(data.FirstName), html.EscapeString(data.LastName)),
		Html:    htmlBody,
		ReplyTo: data.Email,
	}

	sentResponse, err := client.Emails.Send(params)
	if err != nil {
		return sentResponse, fmt.Errorf("fehler beim Senden der E-Mail: %w", err)
	}

	if sentResponse != nil {
		log.Println("Utils: Email successfully sent")
	}
	return sentResponse, nil
}
