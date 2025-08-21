package routes

import (
	"modern-social-media/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, d Deps) {
	rg.POST("/auth/register", handlers.Register(d.Models.Users, d.Models.VerificationCodes, d.Mailer, d.JWTSecret))
	rg.POST("/auth/login", handlers.Login(d.Models.Users, d.Models.VerificationCodes, d.Mailer, d.JWTSecret, d.Email2FAEnabled))
	rg.POST("/auth/refresh", handlers.Refresh(d.JWTSecret))
	rg.POST("/auth/verify-email", handlers.VerifyEmail(d.Models.Users, d.Models.VerificationCodes))
	rg.POST("/auth/resend-verify-email", handlers.ResendVerificationEmail(d.Models.Users, d.Models.VerificationCodes, d.Mailer))
	rg.POST("/auth/2fa/verify", handlers.VerifyLogin2FA(d.Models.Users, d.Models.VerificationCodes, d.JWTSecret))
	rg.POST("/auth/2fa/request", handlers.Request2FACode(d.Models.Users, d.Models.VerificationCodes, d.Mailer, d.Email2FAEnabled))
	rg.POST("/auth/toggle-2fa", handlers.Toggle2FA(d.Models.Users, d.JWTSecret))
}
