package routes

import (
	"modern-social-media/internal/repository"
	"modern-social-media/internal/services"
)

type Deps struct {
	Models          repository.Models
	Mailer          services.EmailSender
	JWTSecret       string
	AdminToken      string
	Email2FAEnabled bool
}
