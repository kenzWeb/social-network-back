package main

import (
	"log"
	"modern-social-media/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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

	r := gin.Default()

	r.Run(":5001")
}
