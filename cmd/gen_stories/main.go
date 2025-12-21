package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"modern-social-media/internal/models"

	"github.com/joho/godotenv"
	"github.com/lucsky/cuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(); err != nil {
			log.Println("Error loading .env file")
		}
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "social_network"),
		getEnv("DB_PORT", "5432"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}

	for _, user := range users {
		numStories := rand.Intn(3) + 1 
		for i := 0; i < numStories; i++ {
			story := models.Story{
				UserID:    user.ID,
				MediaURL:  fmt.Sprintf("https://picsum.photos/1080/1920?random=%s", cuid.New()),
				MediaType: "image",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := db.Create(&story).Error; err != nil {
				log.Printf("Failed to create story for user %s: %v", user.Username, err)
				continue
			}
			fmt.Printf("Created story for user: %s\n", user.Username)
		}
	}

	fmt.Println("Story seeding completed successfully")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
