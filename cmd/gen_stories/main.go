package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
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

	downloadDir := filepath.Join("uploads", "stories", "random")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		log.Fatalf("Failed to create directory %s: %v", downloadDir, err)
	}

	log.Println("Creating stories with images locally...")

	for _, user := range users {
		numStories := rand.Intn(3) + 1
		for i := 0; i < numStories; i++ {
			seed := cuid.New()
			remoteURL := fmt.Sprintf("https://picsum.photos/seed/%s/1080/1920", seed)
			filename := fmt.Sprintf("%s.jpg", seed)
			localPath := filepath.Join(downloadDir, filename)
			dbURL := ""

			if err := downloadImage(remoteURL, localPath); err != nil {
				log.Printf("Failed to download image for story %d: %v. Leaving image blank.", i+1, err)
			} else {
				dbURL = "/uploads/stories/random/" + filename
			}

			story := models.Story{
				UserID:    user.ID,
				MediaURL:  dbURL,
				MediaType: "image",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := db.Create(&story).Error; err != nil {
				log.Printf("Failed to create story for user %s: %v", user.Username, err)
				continue
			}
			fmt.Printf("Created story for user: %s (Image: %s)\n", user.Username, dbURL)
		}
	}

	fmt.Println("Story seeding completed successfully")
}

func downloadImage(url, destPath string) error {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "modern-social-media-seeder/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
