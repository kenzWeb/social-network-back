package main

import (
	"fmt"
	"log"
	"modern-social-media/internal/models"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load env
	if err := godotenv.Load("../../.env"); err != nil {
		godotenv.Load()
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

	var count int64
	if err := db.Model(&models.Follow{}).Count(&count).Error; err != nil {
		log.Fatalf("Failed to count follows: %v", err)
	}

	fmt.Printf("Total Follows in DB: %d\n", count)

	var follows []models.Follow
	if err := db.Limit(5).Find(&follows).Error; err != nil {
		log.Fatalf("Failed to find follows: %v", err)
	}

	for _, f := range follows {
		fmt.Printf("Follow: %s -> %s (Created: %v)\n", f.FollowerID, f.FollowingID, f.CreatedAt)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
