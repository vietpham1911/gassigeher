package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// EmailService handles sending emails via Gmail API
type EmailService struct {
	service   *gmail.Service
	fromEmail string
}

// NewEmailService creates a new email service
func NewEmailService(clientID, clientSecret, refreshToken, fromEmail string) (*EmailService, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.GmailSendScope},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}

	client := config.Client(oauth2.NoContext, token)

	service, err := gmail.NewService(oauth2.NoContext, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return &EmailService{
		service:   service,
		fromEmail: fromEmail,
	}, nil
}

// SendEmail sends an email
func (s *EmailService) SendEmail(to, subject, body string) error {
	var message gmail.Message

	emailContent := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		"%s", s.fromEmail, to, subject, body)

	message.Raw = base64.URLEncoding.EncodeToString([]byte(emailContent))

	_, err := s.service.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent to %s: %s", to, subject)
	return nil
}

// SendVerificationEmail sends an email verification link
func (s *EmailService) SendVerificationEmail(to, name, token string) error {
	subject := "Willkommen bei Gassigeher - E-Mail-Adresse bestätigen"

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Titillium, Arial, sans-serif; line-height: 1.6; color: #26272b; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #82b965; color: white; padding: 20px; text-align: center; border-radius: 6px 6px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 6px 6px; }
        .button { display: inline-block; padding: 12px 30px; background-color: #82b965; color: white; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🐕 Willkommen bei Gassigeher</h1>
        </div>
        <div class="content">
            <p>Hallo {{.Name}},</p>
            <p>vielen Dank für Ihre Registrierung bei Gassigeher! Bitte bestätigen Sie Ihre E-Mail-Adresse, um Ihr Konto zu aktivieren.</p>
            <p style="text-align: center;">
                <a href="http://localhost:8080/verify?token={{.Token}}" class="button">E-Mail-Adresse bestätigen</a>
            </p>
            <p>Oder kopieren Sie diesen Link in Ihren Browser:</p>
            <p style="word-break: break-all; font-size: 12px; color: #666;">
                http://localhost:8080/verify?token={{.Token}}
            </p>
            <p>Dieser Link ist 24 Stunden gültig.</p>
            <p>Wenn Sie sich nicht bei Gassigeher registriert haben, können Sie diese E-Mail ignorieren.</p>
        </div>
        <div class="footer">
            <p>© 2025 Gassigeher. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>
`

	t := template.Must(template.New("verification").Parse(tmpl))
	var body bytes.Buffer
	if err := t.Execute(&body, map[string]string{"Name": name, "Token": token}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}

// SendWelcomeEmail sends a welcome email after verification
func (s *EmailService) SendWelcomeEmail(to, name string) error {
	subject := "Los geht's! Ihr Konto ist aktiviert"

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Titillium, Arial, sans-serif; line-height: 1.6; color: #26272b; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #82b965; color: white; padding: 20px; text-align: center; border-radius: 6px 6px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 6px 6px; }
        .feature { margin: 15px 0; padding: 15px; background-color: white; border-left: 4px solid #82b965; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Willkommen bei Gassigeher!</h1>
        </div>
        <div class="content">
            <p>Hallo {{.Name}},</p>
            <p>Ihr Konto ist jetzt aktiviert! Sie können sofort mit dem Buchen von Hunden beginnen.</p>

            <h3>So funktioniert's:</h3>

            <div class="feature">
                <strong>🐶 Hunde durchsuchen</strong><br>
                Sehen Sie sich alle verfügbaren Hunde an und filtern Sie nach Größe, Rasse und Erfahrungslevel.
            </div>

            <div class="feature">
                <strong>📅 Termine buchen</strong><br>
                Wählen Sie einen Hund und einen Zeitpunkt für Ihren Spaziergang. Sie können die vorgeschlagenen Zeiten anpassen.
            </div>

            <div class="feature">
                <strong>⭐ Erfahrungslevel</strong><br>
                Sie starten als "Grün" (Anfänger). Sie können höhere Levels beantragen, um Zugang zu anspruchsvolleren Hunden zu erhalten:
                <ul>
                    <li><strong>Grün:</strong> Alle Anfänger (Standard)</li>
                    <li><strong>Blau:</strong> Erfahrene Gassigeher</li>
                    <li><strong>Orange:</strong> Nur erfahrene Gassigeher</li>
                </ul>
            </div>

            <p>Bei Fragen oder Problemen wenden Sie sich bitte an unseren Support.</p>

            <p style="text-align: center; margin-top: 30px;">
                <a href="http://localhost:8080" style="display: inline-block; padding: 12px 30px; background-color: #82b965; color: white; text-decoration: none; border-radius: 6px;">Zur Anwendung</a>
            </p>
        </div>
        <div class="footer">
            <p>© 2025 Gassigeher. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>
`

	t := template.Must(template.New("welcome").Parse(tmpl))
	var body bytes.Buffer
	if err := t.Execute(&body, map[string]string{"Name": name}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}

// SendPasswordResetEmail sends a password reset link
func (s *EmailService) SendPasswordResetEmail(to, name, token string) error {
	subject := "Passwort zurücksetzen - Gassigeher"

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Titillium, Arial, sans-serif; line-height: 1.6; color: #26272b; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #82b965; color: white; padding: 20px; text-align: center; border-radius: 6px 6px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 6px 6px; }
        .button { display: inline-block; padding: 12px 30px; background-color: #82b965; color: white; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .warning { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔑 Passwort zurücksetzen</h1>
        </div>
        <div class="content">
            <p>Hallo {{.Name}},</p>
            <p>Sie haben eine Anfrage zum Zurücksetzen Ihres Passworts gestellt. Klicken Sie auf den Button unten, um ein neues Passwort festzulegen.</p>
            <p style="text-align: center;">
                <a href="http://localhost:8080/reset-password?token={{.Token}}" class="button">Neues Passwort festlegen</a>
            </p>
            <p>Oder kopieren Sie diesen Link in Ihren Browser:</p>
            <p style="word-break: break-all; font-size: 12px; color: #666;">
                http://localhost:8080/reset-password?token={{.Token}}
            </p>
            <div class="warning">
                <strong>⚠️ Wichtig:</strong> Dieser Link ist nur 1 Stunde gültig.
            </div>
            <p>Wenn Sie diese Anfrage nicht gestellt haben, können Sie diese E-Mail ignorieren. Ihr Passwort bleibt unverändert.</p>
        </div>
        <div class="footer">
            <p>© 2025 Gassigeher. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>
`

	t := template.Must(template.New("reset").Parse(tmpl))
	var body bytes.Buffer
	if err := t.Execute(&body, map[string]string{"Name": name, "Token": token}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}

// SendBookingConfirmation sends a booking confirmation email
func (s *EmailService) SendBookingConfirmation(to, name, dogName, date, walkType, scheduledTime string) error {
	subject := fmt.Sprintf("Buchungsbestätigung - %s", dogName)

	walkTypeLabel := "Morgen"
	if walkType == "evening" {
		walkTypeLabel = "Abend"
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #26272b; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #82b965; color: white; padding: 20px; text-align: center; border-radius: 6px 6px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 6px 6px; }
        .booking-details { background-color: white; padding: 20px; margin: 20px 0; border-radius: 6px; border-left: 4px solid #82b965; }
        .detail-row { margin: 10px 0; }
        .label { font-weight: 600; color: #666; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Buchung bestätigt!</h1>
        </div>
        <div class="content">
            <p>Hallo {{.Name}},</p>
            <p>Ihre Buchung wurde erfolgreich bestätigt.</p>

            <div class="booking-details">
                <h3 style="margin-top: 0;">Buchungsdetails</h3>
                <div class="detail-row">
                    <span class="label">Hund:</span> {{.DogName}}
                </div>
                <div class="detail-row">
                    <span class="label">Datum:</span> {{.Date}}
                </div>
                <div class="detail-row">
                    <span class="label">Spaziergang:</span> {{.WalkType}}
                </div>
                <div class="detail-row">
                    <span class="label">Uhrzeit:</span> {{.ScheduledTime}} Uhr
                </div>
            </div>

            <p>Sie erhalten eine Erinnerung 1 Stunde vor Ihrem Spaziergang.</p>
            <p>Falls Sie den Termin stornieren möchten, tun Sie dies bitte mindestens 12 Stunden im Voraus über Ihr Dashboard.</p>
        </div>
        <div class="footer">
            <p>© 2025 Gassigeher. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>
`

	t := template.Must(template.New("booking").Parse(tmpl))
	var body bytes.Buffer
	data := map[string]string{
		"Name":          name,
		"DogName":       dogName,
		"Date":          date,
		"WalkType":      walkTypeLabel,
		"ScheduledTime": scheduledTime,
	}
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}

// SendBookingCancellation sends a booking cancellation confirmation (user-initiated)
func (s *EmailService) SendBookingCancellation(to, name, dogName, date, walkType string) error {
	subject := fmt.Sprintf("Buchung storniert - %s", dogName)

	walkTypeLabel := "Morgen"
	if walkType == "evening" {
		walkTypeLabel = "Abend"
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #26272b; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; border-radius: 6px 6px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 6px 6px; }
        .booking-details { background-color: white; padding: 20px; margin: 20px 0; border-radius: 6px; border-left: 4px solid #dc3545; }
        .detail-row { margin: 10px 0; }
        .label { font-weight: 600; color: #666; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Buchung storniert</h1>
        </div>
        <div class="content">
            <p>Hallo {{.Name}},</p>
            <p>Ihre Buchung wurde erfolgreich storniert.</p>

            <div class="booking-details">
                <h3 style="margin-top: 0;">Stornierte Buchung</h3>
                <div class="detail-row">
                    <span class="label">Hund:</span> {{.DogName}}
                </div>
                <div class="detail-row">
                    <span class="label">Datum:</span> {{.Date}}
                </div>
                <div class="detail-row">
                    <span class="label">Spaziergang:</span> {{.WalkType}}
                </div>
            </div>

            <p>Sie können jederzeit eine neue Buchung vornehmen.</p>
        </div>
        <div class="footer">
            <p>© 2025 Gassigeher. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>
`

	t := template.Must(template.New("cancellation").Parse(tmpl))
	var body bytes.Buffer
	data := map[string]string{
		"Name":     name,
		"DogName":  dogName,
		"Date":     date,
		"WalkType": walkTypeLabel,
	}
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}

// SendAdminCancellation sends an admin cancellation notification
func (s *EmailService) SendAdminCancellation(to, name, dogName, date, walkType, reason string) error {
	subject := fmt.Sprintf("Deine Buchung wurde storniert - %s", dogName)

	walkTypeLabel := "Morgen"
	if walkType == "evening" {
		walkTypeLabel = "Abend"
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #26272b; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; border-radius: 6px 6px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 6px 6px; }
        .booking-details { background-color: white; padding: 20px; margin: 20px 0; border-radius: 6px; border-left: 4px solid #dc3545; }
        .reason-box { background-color: #fff3cd; padding: 15px; margin: 20px 0; border-radius: 6px; border-left: 4px solid #ffc107; }
        .detail-row { margin: 10px 0; }
        .label { font-weight: 600; color: #666; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Buchung storniert</h1>
        </div>
        <div class="content">
            <p>Hallo {{.Name}},</p>
            <p>Leider mussten wir Ihre folgende Buchung stornieren:</p>

            <div class="booking-details">
                <h3 style="margin-top: 0;">Stornierte Buchung</h3>
                <div class="detail-row">
                    <span class="label">Hund:</span> {{.DogName}}
                </div>
                <div class="detail-row">
                    <span class="label">Datum:</span> {{.Date}}
                </div>
                <div class="detail-row">
                    <span class="label">Spaziergang:</span> {{.WalkType}}
                </div>
            </div>

            <div class="reason-box">
                <strong>Grund der Stornierung:</strong><br>
                {{.Reason}}
            </div>

            <p>Wir entschuldigen uns für die Unannehmlichkeiten. Sie können gerne einen anderen Termin buchen.</p>
        </div>
        <div class="footer">
            <p>© 2025 Gassigeher. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>
`

	t := template.Must(template.New("admin_cancel").Parse(tmpl))
	var body bytes.Buffer
	data := map[string]string{
		"Name":     name,
		"DogName":  dogName,
		"Date":     date,
		"WalkType": walkTypeLabel,
		"Reason":   reason,
	}
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}

// SendBookingReminder sends a reminder 1 hour before the booking
func (s *EmailService) SendBookingReminder(to, name, dogName, date, walkType, scheduledTime string) error {
	subject := fmt.Sprintf("Erinnerung: Gassirunde mit %s in 1 Stunde", dogName)

	walkTypeLabel := "Morgen"
	if walkType == "evening" {
		walkTypeLabel = "Abend"
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #26272b; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #17a2b8; color: white; padding: 20px; text-align: center; border-radius: 6px 6px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 6px 6px; }
        .booking-details { background-color: white; padding: 20px; margin: 20px 0; border-radius: 6px; border-left: 4px solid #17a2b8; }
        .detail-row { margin: 10px 0; }
        .label { font-weight: 600; color: #666; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔔 Erinnerung</h1>
        </div>
        <div class="content">
            <p>Hallo {{.Name}},</p>
            <p>Dies ist eine Erinnerung an Ihren bevorstehenden Spaziergang:</p>

            <div class="booking-details">
                <h3 style="margin-top: 0;">Ihr Spaziergang</h3>
                <div class="detail-row">
                    <span class="label">Hund:</span> {{.DogName}}
                </div>
                <div class="detail-row">
                    <span class="label">Datum:</span> {{.Date}}
                </div>
                <div class="detail-row">
                    <span class="label">Spaziergang:</span> {{.WalkType}}
                </div>
                <div class="detail-row">
                    <span class="label">Uhrzeit:</span> {{.ScheduledTime}} Uhr
                </div>
            </div>

            <p>Viel Spaß beim Spaziergang!</p>
        </div>
        <div class="footer">
            <p>© 2025 Gassigeher. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>
`

	t := template.Must(template.New("reminder").Parse(tmpl))
	var body bytes.Buffer
	data := map[string]string{
		"Name":          name,
		"DogName":       dogName,
		"Date":          date,
		"WalkType":      walkTypeLabel,
		"ScheduledTime": scheduledTime,
	}
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}

// SendBookingMoved sends an email when admin moves a booking
func (s *EmailService) SendBookingMoved(to, name, dogName, oldDate, oldWalkType, oldTime, newDate, newWalkType, newTime, reason string) error {
	subject := fmt.Sprintf("Deine Buchung wurde verschoben - %s", dogName)

	oldWalkLabel := "Morgen"
	if oldWalkType == "evening" {
		oldWalkLabel = "Abend"
	}

	newWalkLabel := "Morgen"
	if newWalkType == "evening" {
		newWalkLabel = "Abend"
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #26272b; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #17a2b8; color: white; padding: 20px; text-align: center; border-radius: 6px 6px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 6px 6px; }
        .booking-details { background-color: white; padding: 20px; margin: 20px 0; border-radius: 6px; }
        .old-details { border-left: 4px solid #dc3545; }
        .new-details { border-left: 4px solid #28a745; margin-top: 20px; }
        .reason-box { background-color: #fff3cd; padding: 15px; margin: 20px 0; border-radius: 6px; border-left: 4px solid #ffc107; }
        .detail-row { margin: 10px 0; }
        .label { font-weight: 600; color: #666; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Buchung verschoben</h1>
        </div>
        <div class="content">
            <p>Hallo {{.Name}},</p>
            <p>Ihre Buchung wurde auf einen neuen Termin verschoben:</p>

            <div class="booking-details old-details">
                <h3 style="margin-top: 0; color: #dc3545;">Alter Termin</h3>
                <div class="detail-row">
                    <span class="label">Hund:</span> {{.DogName}}
                </div>
                <div class="detail-row">
                    <span class="label">Datum:</span> {{.OldDate}}
                </div>
                <div class="detail-row">
                    <span class="label">Spaziergang:</span> {{.OldWalkType}}
                </div>
                <div class="detail-row">
                    <span class="label">Uhrzeit:</span> {{.OldTime}} Uhr
                </div>
            </div>

            <div class="booking-details new-details">
                <h3 style="margin-top: 0; color: #28a745;">Neuer Termin</h3>
                <div class="detail-row">
                    <span class="label">Hund:</span> {{.DogName}}
                </div>
                <div class="detail-row">
                    <span class="label">Datum:</span> {{.NewDate}}
                </div>
                <div class="detail-row">
                    <span class="label">Spaziergang:</span> {{.NewWalkType}}
                </div>
                <div class="detail-row">
                    <span class="label">Uhrzeit:</span> {{.NewTime}} Uhr
                </div>
            </div>

            <div class="reason-box">
                <strong>Grund der Verschiebung:</strong><br>
                {{.Reason}}
            </div>

            <p>Wir entschuldigen uns für die Unannehmlichkeiten. Bei Fragen oder Problemen wenden Sie sich bitte an uns.</p>
        </div>
        <div class="footer">
            <p>© 2025 Gassigeher. Alle Rechte vorbehalten.</p>
        </div>
    </div>
</body>
</html>
`

	t := template.Must(template.New("moved").Parse(tmpl))
	var body bytes.Buffer
	data := map[string]string{
		"Name":        name,
		"DogName":     dogName,
		"OldDate":     oldDate,
		"OldWalkType": oldWalkLabel,
		"OldTime":     oldTime,
		"NewDate":     newDate,
		"NewWalkType": newWalkLabel,
		"NewTime":     newTime,
		"Reason":      reason,
	}
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}
