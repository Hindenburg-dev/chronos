package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"chronos/internal/codeforces"
	"chronos/internal/config"
)

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username: username, password: password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:", "username:", "VXNlcm5hbWU6":
			return []byte(a.username), nil
		case "Password:", "password:", "UGFzc3dvcmQ6":
			return []byte(a.password), nil
		default:
			return []byte(a.password), nil
		}
	}
	return nil, nil
}

// SendPlan 将指定日期的排程结果发送到配置的目标邮箱
func SendPlan(cfg *config.EmailConfig, date, planContent string, cfProblems ...codeforces.ProblemDetail) error {
	if cfg == nil || !cfg.Enabled {
		return nil
	}

	from := strings.TrimSpace(cfg.From)
	password := strings.TrimSpace(cfg.Password)
	smtpHost := strings.TrimSpace(cfg.SMTPHost)
	to := strings.TrimSpace(cfg.To)

	if smtpHost == "" || from == "" || password == "" {
		return fmt.Errorf("邮件配置不完整 (缺少 smtp_host, from 或 password)")
	}

	if to == "" {
		to = from // 默认发送给发件人自己
	}

	port := cfg.SMTPPort
	if port == 0 {
		port = 465 // 默认 465 SSL
	}

	subject := fmt.Sprintf("[Chronos 日程] %s 执行计划表", date)
	htmlBody := MarkdownToHTML(date, planContent, cfProblems...)

	// 构建标准 MIME 邮件报文 (HTML 格式)
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var messageBuilder strings.Builder
	for k, v := range headers {
		messageBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	messageBuilder.WriteString("\r\n")
	messageBuilder.WriteString(htmlBody)

	rawMessage := []byte(messageBuilder.String())
	addr := fmt.Sprintf("%s:%d", smtpHost, port)

	// 针对 465 端口（QQ/163 邮箱常用的 SSL/TLS 直连）
	if port == 465 {
		tlsConfig := &tls.Config{
			ServerName: smtpHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("连接 SMTP 服务器失败: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, smtpHost)
		if err != nil {
			return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
		}
		defer client.Close()

		// 优先尝试 LoginAuth (国内 163/QQ 邮箱的主流认证方式)
		auth := LoginAuth(from, password)
		if err = client.Auth(auth); err != nil {
			// 如果 LoginAuth 失败，回退尝试 PlainAuth
			plainAuth := smtp.PlainAuth("", from, password, smtpHost)
			if errPlain := client.Auth(plainAuth); errPlain != nil {
				return fmt.Errorf("SMTP 身份验证失败 (请检查邮箱与授权码): %w", err)
			}
		}

		if err = client.Mail(from); err != nil {
			return fmt.Errorf("设置发件人失败: %w", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("获取数据写入流失败: %w", err)
		}

		if _, err = w.Write(rawMessage); err != nil {
			return fmt.Errorf("写入邮件内容失败: %w", err)
		}

		if err = w.Close(); err != nil {
			return fmt.Errorf("关闭数据流失败: %w", err)
		}

		return client.Quit()
	}

	// 针对 587/25 等 STARTTLS / 标准端口
	auth := LoginAuth(from, password)
	return smtp.SendMail(addr, auth, from, []string{to}, rawMessage)
}
