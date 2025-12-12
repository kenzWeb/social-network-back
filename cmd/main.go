package main

import (
	"context"
	"log"
	"time"
	"modern-social-media/internal/env"
	"modern-social-media/internal/repository"
	"modern-social-media/internal/services"

	"github.com/joho/godotenv"
)

type application struct {
	port            int
	jwtSecret       string
	adminToken      string
	models          repository.Models
	mailer          services.EmailSender
	email2FAEnabled bool
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("Не удалось загрузить .env файл: %v", err)
	}

	db, err := repository.InitDb()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Ошибка получения sql.DB: %v", err)
	}
	defer sqlDB.Close()

	err = repository.MigrateDatabase(db)
	if err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	log.Println("Успешно подключено к PostgreSQL базе данных с GORM")
	log.Println("Миграции выполнены успешно")

	models := repository.NewModels(db)
	
	storyService := services.NewStoryService(models.Stories)
	go func() {
		ctx := context.Background()
		if err := storyService.CleanupExpiredStories(ctx); err != nil {
			log.Printf("Initial story cleanup failed: %v", err)
		}
		
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := storyService.CleanupExpiredStories(ctx); err != nil {
				log.Printf("Story cleanup failed: %v", err)
			}
		}
	}()
	mailer := &services.SMTPSender{
		Host:     env.GetEnvString("SMTP_HOST", "localhost"),
		Port:     env.GetEnvInt("SMTP_PORT", 587),
		Username: env.GetEnvString("SMTP_USERNAME", ""),
		Password: env.GetEnvString("SMTP_PASSWORD", ""),
		From:     env.GetEnvString("SMTP_FROM", "noreply@example.com"),
	}
	app := &application{
		port:            env.GetEnvInt("PORT", 8080),
		jwtSecret:       env.GetEnvString("JWT_SECRET", ""),
		adminToken:      env.GetEnvString("ADMIN_TOKEN", ""),
		models:          *models,
		mailer:          mailer,
		email2FAEnabled: env.GetEnvBool("EMAIL_2FA_ENABLED", true),
	}

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}
