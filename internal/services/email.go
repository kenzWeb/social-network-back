package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"
)

type EmailSender interface {
	Send(to, subject, body string) error
}

type SMTPSender struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func (s *SMTPSender) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	htmlBody := `
<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

body {
	background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
	padding: 40px 20px;
	min-height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
}

.email-wrapper {
	width: 100%;
	max-width: 600px;
	margin: 0 auto;
}

.email-container {
	background: #282828;
	border-radius: 16px;
	overflow: hidden;
	box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
	border: 1px solid rgba(255, 253, 2, 0.2);
}

.email-header {
	background: linear-gradient(135deg, #666666 0%, #000000 100%);
	padding: 40px 30px;
	text-align: center;
	position: relative;
}

.email-header::before {
	content: '';
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	height: 3px;
	background: linear-gradient(90deg, #fffd02, #f0f073, #fffd02);
}

.logo-text {
	font-size: 32px;
	font-weight: 700;
	color: #fffd02;
	letter-spacing: -0.5px;
	margin-bottom: 8px;
}

.tagline {
	color: #999;
	font-size: 14px;
	letter-spacing: 0.5px;
}

.email-body {
	padding: 40px 30px;
}

.greeting {
	font-size: 20px;
	color: #f0f0f0;
	margin-bottom: 20px;
	font-weight: 500;
}

.message {
	color: #b0b0b0;
	line-height: 1.8;
	margin-bottom: 30px;
	font-size: 15px;
}

.code-section {
	background: linear-gradient(135deg, rgba(255, 253, 2, 0.05) 0%, rgba(255, 253, 2, 0.1) 100%);
	border: 2px solid #fffd02;
	border-radius: 12px;
	padding: 30px;
	margin: 30px 0;
	text-align: center;
	position: relative;
	overflow: hidden;
}

.code-section::before {
	content: '';
	position: absolute;
	top: -50%;
	left: -50%;
	width: 200%;
	height: 200%;
	background: radial-gradient(circle, rgba(255, 253, 2, 0.1) 0%, transparent 70%);
	animation: pulse 3s ease-in-out infinite;
}

@keyframes pulse {
	0%, 100% { transform: scale(1); opacity: 0.5; }
	50% { transform: scale(1.1); opacity: 0.8; }
}

.code-label {
	font-size: 12px;
	text-transform: uppercase;
	letter-spacing: 1.5px;
	color: #888;
	margin-bottom: 12px;
	font-weight: 600;
}

.code-value {
	font-size: 42px;
	font-weight: 700;
	color: #fffd02;
	letter-spacing: 8px;
	font-family: 'Courier New', monospace;
	position: relative;
	text-shadow: 0 0 20px rgba(255, 253, 2, 0.3);
}

.code-hint {
	font-size: 13px;
	color: #666;
	margin-top: 12px;
}

.warning {
	background: rgba(255, 100, 100, 0.1);
	border-left: 3px solid #ff6464;
	padding: 15px 20px;
	border-radius: 6px;
	color: #ffb3b3;
	font-size: 14px;
	margin-top: 25px;
}

.email-footer {
	background: #1f1f1f;
	padding: 30px;
	text-align: center;
	border-top: 1px solid rgba(255, 253, 2, 0.1);
}

.footer-links {
	margin-bottom: 20px;
}

.footer-link {
	color: #888;
	text-decoration: none;
	font-size: 13px;
	margin: 0 15px;
	transition: color 0.3s;
}

.footer-link:hover {
	color: #fffd02;
}

.copyright {
	color: #555;
	font-size: 12px;
	margin-top: 15px;
}

.divider {
	height: 1px;
	background: linear-gradient(90deg, transparent, rgba(255, 253, 2, 0.3), transparent);
	margin: 25px 0;
}

@media only screen and (max-width: 600px) {
	body {
		padding: 20px 10px;
	}
	
	.email-header, .email-body, .email-footer {
		padding: 25px 20px;
	}
	
	.code-value {
		font-size: 32px;
		letter-spacing: 4px;
	}
}
</style>
</head>
<body>
	<div class="email-wrapper">
		<div class="email-container">
			<div class="email-header">
				<div class="logo-text">Modern Social</div>
				<div class="tagline">CONNECTING THE FUTURE</div>
			</div>
			
			<div class="email-body">
				<div class="greeting">Здравствуйте! 👋</div>
				
				<p class="message">
					Вы получили это письмо, потому что был запрошен код подтверждения для вашего аккаунта. 
					Используйте код ниже для завершения процесса.
				</p>
				
				<div class="code-section">
					<div class="code-label">Ваш код подтверждения</div>
					<div class="code-value">` + body + `</div>
					<div class="code-hint">Код действителен в течение 15 минут</div>
				</div>
				
				<div class="divider"></div>
				
				<p class="message">
					Если вы не запрашивали этот код, просто проигнорируйте это письмо. 
					Ваша безопасность - наш приоритет.
				</p>
				
				<div class="warning">
					⚠️ Никогда не передавайте этот код третьим лицам
				</div>
			</div>
			
			<div class="email-footer">
				<div class="footer-links">
					<a href="#" class="footer-link">Помощь</a>
					<a href="#" class="footer-link">Политика</a>
					<a href="#" class="footer-link">Контакты</a>
				</div>
				<div class="copyright">
					© ` + fmt.Sprintf("%d", time.Now().Year()) + ` Modern Social Network. Все права защищены.
				</div>
			</div>
		</div>
	</div>
</body>
</html>
`

	msg := []byte("From: " + s.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" + htmlBody)

	if s.Port == 465 {
	}
	return smtp.SendMail(addr, auth, s.From, []string{to}, msg)
}

func tlsConfig(host string) *tls.Config {
	return &tls.Config{ServerName: host}
}
