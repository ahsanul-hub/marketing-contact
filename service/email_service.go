package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"app/config"
	"app/dto/model"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

type EmailConfig struct {
	SMTPHost   string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	FromEmail  string
	FromName   string
	ToEmails   []string
	Subject    string
	Body       string
	Attachment []byte
	FileName   string
}

type MerchantEmailConfig struct {
	ClientName string
	AppID      string
	EmailTo    string
}

func (es *EmailService) SendTransactionReport(transactions []model.Transactions, merchantName string, startDate, endDate time.Time) error {
	// Generate Excel report
	excelData, err := GenerateExcelReport(transactions, merchantName)
	if err != nil {
		return fmt.Errorf("error generating Excel report: %v", err)
	}

	// Email configuration - support untuk webmail
	emailConfig := EmailConfig{
		SMTPHost:   config.Config("SMTP_HOST", ""),
		SMTPPort:   config.Config("SMTP_PORT", "587"),
		SMTPUser:   config.Config("SMTP_USER", ""),
		SMTPPass:   config.Config("SMTP_PASS", ""),
		FromEmail:  config.Config("FROM_EMAIL", ""),
		FromName:   config.Config("FROM_NAME", "Reconcile Redision"),
		ToEmails:   strings.Split(config.Config("TO_EMAILS", "zhangshijun@ushareit.com,rinkesh.sharma@payermax.com,fanny@redision.com,juan.rivaldo@ushareit.com,payment_report@payermax.com,huzhihong@payermax.com,huzhihong@ushareit.com,chengjiexin@ushareit.com"), ","),
		Subject:    "Transaction Bill Detail",
		Body:       es.generateEmailBody(endDate),
		Attachment: excelData,
		FileName:   fmt.Sprintf("PM Max %s.xlsx", startDate.AddDate(0, 0, 1).Format("02 January 2006")),
	}

	// Validate email config
	if emailConfig.SMTPHost == "" || emailConfig.SMTPUser == "" || emailConfig.SMTPPass == "" || emailConfig.FromEmail == "" || len(emailConfig.ToEmails) == 0 {
		return fmt.Errorf("email configuration incomplete: SMTP_HOST, SMTP_USER, SMTP_PASS, FROM_EMAIL, or TO_EMAILS not set")
	}

	// Log email configuration (tanpa password)
	// log.Printf("Sending email via %s:%s", emailConfig.SMTPHost, emailConfig.SMTPPort)
	// log.Printf("From: %s <%s>", emailConfig.FromName, emailConfig.FromEmail)
	// log.Printf("To: %s", strings.Join(emailConfig.ToEmails, ", "))
	// log.Printf("Subject: %s", emailConfig.Subject)
	// log.Printf("Attachment: %s (%d bytes)", emailConfig.FileName, len(emailConfig.Attachment))

	// Send email
	err = es.sendEmail(emailConfig)
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	log.Printf("Successfully sent transaction report email for merchant %s: %s", merchantName, emailConfig.FileName)
	return nil
}

func (es *EmailService) generateEmailBody(endDate time.Time) string {
	body := fmt.Sprintf(`
Dear Partner,

Here i attached your transaction bill period %s.

Regards,
Fanny
`, endDate.Format("02 January 2006"))

	return body
}

func (es *EmailService) sendEmail(config EmailConfig) error {
	// SMTP authentication
	auth := smtp.PlainAuth("", config.SMTPUser, config.SMTPPass, config.SMTPHost)

	// Generate boundary
	boundary := "boundary123456789"

	// Email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail)
	headers["To"] = strings.Join(config.ToEmails, ",")
	headers["Subject"] = config.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%s", boundary)

	// Build email message
	var message bytes.Buffer

	// Write headers
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")

	// Write text part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	message.WriteString("\r\n")
	message.WriteString(config.Body)
	message.WriteString("\r\n")

	// Write attachment part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet\r\n")
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", config.FileName))
	message.WriteString("\r\n")

	// Encode attachment to base64
	encodedAttachment := base64.StdEncoding.EncodeToString(config.Attachment)

	// Split base64 into lines of 76 characters
	for i := 0; i < len(encodedAttachment); i += 76 {
		end := i + 76
		if end > len(encodedAttachment) {
			end = len(encodedAttachment)
		}
		message.WriteString(encodedAttachment[i:end])
		message.WriteString("\r\n")
	}

	// End boundary
	message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// Send email
	addr := fmt.Sprintf("%s:%s", config.SMTPHost, config.SMTPPort)
	err := smtp.SendMail(addr, auth, config.FromEmail, config.ToEmails, message.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
