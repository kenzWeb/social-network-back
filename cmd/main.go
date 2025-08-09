package main

import (
	"log"
	"modern-social-media/internal/env"
	"modern-social-media/internal/repository"

	"github.com/joho/godotenv"
)

type application struct {
	port int
	jwtSecret string
	models repository.Models
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
	app := &application{
		port: env.GetEnvInt("PORT", 8080),
		jwtSecret: env.GetEnvString("JWT_SECRET", "123456"),
		models: *models,
	}

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}
