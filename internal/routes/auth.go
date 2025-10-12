package routes

import (
	"modern-social-media/internal/handlers"
	"modern-social-media/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, d Deps) {
	// Единый экземпляр AuthService для всех эндпоинтов
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

	rg.POST("/auth/register", handlers.RegisterWithService(svc))
	rg.POST("/auth/login", handlers.LoginWithService(svc))
	rg.POST("/auth/refresh", handlers.Refresh(d.JWTSecret))
	rg.POST("/auth/verify-email", handlers.VerifyEmailWithService(svc))
	rg.POST("/auth/resend-verify-email", handlers.ResendVerificationEmailWithService(svc))
	rg.POST("/auth/2fa/verify", handlers.VerifyLogin2FAWithService(svc))
	rg.POST("/auth/2fa/request", handlers.Request2FACodeWithService(svc))
	rg.POST("/auth/toggle-2fa", handlers.Toggle2FAWithService(svc, d.JWTSecret))
}
