package routes

import (
	authhandlers "modern-social-media/internal/handlers/auth"
	"modern-social-media/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, d Deps) {
	svc := services.AuthService{
		Users:           d.Models.Users,
		Codes:           d.Models.VerificationCodes,
		Tokens:          &services.JWTTokenService{Secret: []byte(d.JWTSecret), AccessTTL: 15 * time.Minute, RefreshTTL: 14 * 24 * time.Hour},
		Mailer:          d.Mailer,
		Hasher:          services.Argon2Hasher{},
		Clock:           services.RealClock{},
		Transact:        services.NoopTxRunner{},
		Email2FAEnabled: d.Email2FAEnabled,
	}

	rg.POST("/auth/register", authhandlers.RegisterWithService(svc))
	rg.POST("/auth/login", authhandlers.LoginWithService(svc))
	rg.POST("/auth/refresh", authhandlers.Refresh(d.JWTSecret))
	rg.POST("/auth/verify-email", authhandlers.VerifyEmailWithService(svc))
	rg.POST("/auth/resend-verify-email", authhandlers.ResendVerificationEmailWithService(svc))
	rg.POST("/auth/2fa/verify", authhandlers.VerifyLogin2FAWithService(svc))
	rg.POST("/auth/2fa/request", authhandlers.Request2FACodeWithService(svc))
	rg.POST("/auth/toggle-2fa", authhandlers.Toggle2FAWithService(svc, d.JWTSecret))
}
