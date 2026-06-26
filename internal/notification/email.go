// Package notification 提供基于 SMTP 的邮件发送能力。
// 默认适配 Gmail（smtp.gmail.com:587 + STARTTLS），也可通过环境变量切换为
// SendGrid、Resend、AWS SES 等其他 SMTP 服务商。
package notification

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Config 保存邮件服务所需的 SMTP 配置。
type Config struct {
	Host                string
	Port                int
	Username            string
	Password            string
	From                string
	FromName            string
	Insecure            bool
	AuthMethod          string
	VerificationBaseURL string
}

// Service 提供邮件发送接口。
type Service struct {
	cfg    Config
	logger *zap.Logger
}

// NewService 创建邮件服务，并在配置不完整时返回错误。
func NewService(cfg Config, logger *zap.Logger) (*Service, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	cfg.AuthMethod = strings.ToUpper(cfg.AuthMethod)
	if cfg.AuthMethod == "" {
		cfg.AuthMethod = "LOGIN"
	}
	return &Service{cfg: cfg, logger: logger}, nil
}

func validateConfig(cfg Config) error {
	if cfg.Host == "" {
		return fmt.Errorf("smtp host is required")
	}
	if cfg.Port == 0 {
		return fmt.Errorf("smtp port is required")
	}
	if cfg.Username == "" {
		return fmt.Errorf("smtp username is required")
	}
	if cfg.Password == "" {
		return fmt.Errorf("smtp password is required")
	}
	if cfg.From == "" {
		return fmt.Errorf("smtp from address is required")
	}
	if cfg.AuthMethod != "" && strings.ToUpper(cfg.AuthMethod) != "LOGIN" && strings.ToUpper(cfg.AuthMethod) != "PLAIN" {
		return fmt.Errorf("smtp auth method must be LOGIN or PLAIN")
	}
	return nil
}

// SendVerificationEmail 发送邮箱验证邮件。
// 邮件内容包含一个可点击的验证链接，格式为 {VerificationBaseURL}/api/v1/auth/verify?token={token}
func (s *Service) SendVerificationEmail(to, token string) error {
	link := buildVerificationLink(s.cfg.VerificationBaseURL, token)

	subject := "Verify your email for T2 Travel Terminal"
	textBody := fmt.Sprintf(
		"Welcome to T2 Travel Terminal!\n\nPlease verify your email address by opening the following link:\n\n%s\n\nThis link will expire in 24 hours.\n",
		link,
	)
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Email Verification</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
  <div style="max-width: 480px; margin: 0 auto; padding: 24px;">
    <h2 style="color: #111;">Welcome to T2 Travel Terminal</h2>
    <p>Please verify your email address by clicking the button below. This link will expire in 24 hours.</p>
    <p style="margin: 24px 0;">
      <a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #0066cc; color: #fff; text-decoration: none; border-radius: 4px;">Verify Email</a>
    </p>
    <p style="font-size: 12px; color: #666;">If the button doesn't work, copy and paste this link into your browser:<br>%s</p>
  </div>
</body>
</html>`, link, link)

	msg, err := s.buildMessage(to, subject, textBody, htmlBody)
	if err != nil {
		return fmt.Errorf("build verification email: %w", err)
	}

	if err := s.send(to, msg); err != nil {
		return fmt.Errorf("send verification email to %s: %w", to, err)
	}

	s.logger.Info("verification email sent", zap.String("to", to), zap.String("link", link))
	return nil
}

func buildVerificationLink(baseURL, token string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/api/v1/auth/verify?token=%s", baseURL, token)
}

func (s *Service) buildMessage(to, subject, textBody, htmlBody string) ([]byte, error) {
	from := s.cfg.From
	if s.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", mimeEncodeWord(s.cfg.FromName), s.cfg.From)
	}

	boundary := "t2-travel-terminal-boundary"
	headers := make(textproto.MIMEHeader)
	headers.Set("From", from)
	headers.Set("To", to)
	headers.Set("Subject", mimeEncodeWord(subject))
	headers.Set("Date", time.Now().Format(time.RFC1123Z))
	headers.Set("Message-ID", fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), s.cfg.Host))
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=\"%s\"", boundary))

	var b strings.Builder
	for key, values := range headers {
		for _, value := range values {
			fmt.Fprintf(&b, "%s: %s\r\n", key, value)
		}
	}
	fmt.Fprint(&b, "\r\n")

	fmt.Fprintf(&b, "--%s\r\n", boundary)
	fmt.Fprint(&b, "Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	fmt.Fprint(&b, "Content-Transfer-Encoding: 8bit\r\n\r\n")
	fmt.Fprintf(&b, "%s\r\n\r\n", textBody)

	fmt.Fprintf(&b, "--%s\r\n", boundary)
	fmt.Fprint(&b, "Content-Type: text/html; charset=\"UTF-8\"\r\n")
	fmt.Fprint(&b, "Content-Transfer-Encoding: 8bit\r\n\r\n")
	fmt.Fprintf(&b, "%s\r\n\r\n", htmlBody)

	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return []byte(b.String()), nil
}

// mimeEncodeWord 对非 ASCII 字符进行 MIME 编码（RFC 2047）。
// 纯 ASCII 字符串保持原样。
func mimeEncodeWord(s string) string {
	if isASCII(s) {
		return s
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(s))
	return fmt.Sprintf("=?UTF-8?B?%s?=", encoded)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

// newAuth 根据配置创建 SMTP 认证器。
// Outlook/Hotmail 默认使用 LOGIN；其他部分服务商可能需要 PLAIN。
func (s *Service) newAuth() (smtp.Auth, error) {
	switch s.cfg.AuthMethod {
	case "PLAIN":
		return smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host), nil
	case "LOGIN", "":
		return loginAuth{username: s.cfg.Username, password: s.cfg.Password}, nil
	default:
		return nil, fmt.Errorf("unsupported smtp auth method: %s", s.cfg.AuthMethod)
	}
}

// loginAuth 实现 SMTP AUTH LOGIN 机制。
// Gmail 等服务商在 STARTTLS 后通常使用 LOGIN 或 PLAIN。
type loginAuth struct {
	username string
	password string
}

func (a loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	prompt := string(fromServer)
	switch {
	case strings.EqualFold(prompt, "Username:"):
		return []byte(a.username), nil
	case strings.EqualFold(prompt, "Password:"):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge during LOGIN auth: %q", prompt)
	}
}

func (s *Service) send(to string, msg []byte) error {
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	tlsConfig := &tls.Config{
		ServerName: s.cfg.Host,
	}
	if s.cfg.Insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("starttls: %w", err)
	}

	auth, err := s.newAuth()
	if err != nil {
		return fmt.Errorf("create smtp auth: %w", err)
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("open data writer: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data writer: %w", err)
	}

	return client.Quit()
}
