// Package notification 是邮件发送能力的占位实现。
// 实际项目中请替换为 SendGrid、Amazon SES、Mailgun 或自建的 SMTP 服务。
package notification

import (
	"fmt"

	"go.uber.org/zap"
)

// Service 提供邮件发送接口。
type Service struct {
	logger *zap.Logger
}

// NewService 创建邮件服务。
func NewService(logger *zap.Logger) *Service {
	return &Service{logger: logger}
}

// SendVerificationEmail 发送邮箱验证邮件。
// 当前版本只在日志中打印验证链接，方便本地调试。
func (s *Service) SendVerificationEmail(email, token string) error {
	link := fmt.Sprintf("https://your-domain.com/api/v1/auth/verify?token=%s", token)
	s.logger.Info("verification email",
		zap.String("to", email),
		zap.String("link", link),
	)
	return nil
}
