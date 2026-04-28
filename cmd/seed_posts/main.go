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

var postContents = []string{
	"Just had the best coffee ever! ☕️",
	"Chilling on the weekend, how describes your mood today?",
	"Sunset vibes 🌅 Nothing beats this view.",
	"Coding all night long... 💻 bugs everywhere, but we keep pushing!",
	"Who else is excited for the upcoming weekend?",
	"Exploring new places, will post more pictures soon! 🌍",
	"Reading a fascinating book today. Highly recommend it! 📚",
	"A little progress each day adds up to big results.",
	"Just finished a great workout! Feeling pumped! 💪",
	"Sometimes you just need to take a break and breathe.",
	"What's everyone listening to right now? Need some new tracks! 🎵",
	"Cooking dinner tonight! Wish me luck 🍳",
	"Nature always brings so much peace. 🌿",
	"Throwback to this amazing trip! Missing the ocean.",
	"Just got a new gadget! Setting it up now 🚀",
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(); err != nil {
			log.Println("Error loading .env file or not found, using default env vars")
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

	log.Println("Clearing existing posts...")
	if err := db.Exec("TRUNCATE TABLE posts CASCADE").Error; err != nil {
		log.Println("Truncate failed, trying Delete:", err)
		if err := db.Where("1 = 1").Delete(&models.Post{}).Error; err != nil {
			log.Printf("Failed to clear posts: %v", err)
		}
	}

	log.Println("Fetching users...")
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}

	if len(users) == 0 {
		log.Fatal("No users found in database! Please run cmd/seeder first.")
	}

	log.Println("Creating 20 random posts with images locally (this may take a few seconds)...")
	
	downloadDir := filepath.Join("uploads", "posts", "random")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		log.Fatalf("Failed to create directory %s: %v", downloadDir, err)
	}

	count := 20
	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]
		content := postContents[rand.Intn(len(postContents))]
		seed := cuid.New()

		log.Printf("Downloading image %d/20...", i+1)
		remoteURL := fmt.Sprintf("https://picsum.photos/seed/%s/800/600", seed)
		filename := fmt.Sprintf("%s.jpg", seed)
		localPath := filepath.Join(downloadDir, filename)
		dbURL := ""

		if err := downloadImage(remoteURL, localPath); err != nil {
			log.Printf("Failed to download image for post %d: %v. Leaving image blank.", i+1, err)
			// fallback - no image
		} else {
			dbURL = "/uploads/posts/random/" + filename
		}

		post := models.Post{
			UserID:    user.ID,
			Content:   content,
			ImageURL:  dbURL,
			CreatedAt: time.Now().Add(-time.Duration(rand.Intn(168)) * time.Hour),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(&post).Error; err != nil {
			log.Printf("Failed to create post %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Created post %d/20 for user: %s (Image: %s)\n", i+1, user.Username, dbURL)
	}

	log.Println("Posts seeding completed successfully!")
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
