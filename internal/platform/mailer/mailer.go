package mailer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
)

type Mailer interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}

type SMTPMailer struct {
	host      string
	port      int
	username  string
	password  string
	fromName  string
	fromEmail string
}

func NewSMTP(host string, port int, username, password, fromName, fromEmail string) *SMTPMailer {
	return &SMTPMailer{
		host: host, port: port,
		username: username, password: password,
		fromName: fromName, fromEmail: fromEmail,
	}
}

func (m *SMTPMailer) Send(_ context.Context, to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	from := fmt.Sprintf("%s <%s>", m.fromName, m.fromEmail)
	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s\r\n", from)
	fmt.Fprintf(&msg, "To: %s\r\n", to)
	fmt.Fprintf(&msg, "Subject: %s\r\n", subject)
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	return smtp.SendMail(addr, auth, m.fromEmail, []string{to}, msg.Bytes())
}

const emailChangeTemplate = `<!DOCTYPE html>                                                                                                                                                                      
  <html>                                                                                                                                                                                                          
  <body style="font-family: Arial, sans-serif; padding: 24px; color: #333;">
      <h2>Confirm Email Change</h2>                                                                                                                                                                                 
      <p>Hi,</p>
      <p>We received a request to change the email address on your Ikamers account to this one.</p>                                                                                                                 
      <p>Use the code below to confirm the change:</p>                                                                                                                                                              
      <p style="font-size: 32px; font-weight: bold; letter-spacing: 8px; text-align: center;
                background: #f4f4f4; padding: 20px; border-radius: 8px; margin: 24px 0;">                                                                                                                           
          {{.OTP}}                                                                                                                                                                                                
      </p>                                                                                                                                                                                                          
      <p>This code expires in <strong>{{.Minutes}} minutes</strong>.</p>                                                                                                                                          
      <p style="color: #888;">If you didn't request this change, you can safely ignore this email.</p>                                                                                                              
      <hr style="border: none; border-top: 1px solid #eee; margin: 24px 0;">
      <p style="color: #888; font-size: 12px;">This is an automated message. Please do not reply.</p>                                                                                                               
  </body>                                                                                                                                                                                                           
  </html>`

func RenderEmailChangeOTP(otp string, minutes int) (string, error) {
	tmpl, err := template.New("email_change").Parse(emailChangeTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]any{
		"OTP":     otp,
		"Minutes": minutes,
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
