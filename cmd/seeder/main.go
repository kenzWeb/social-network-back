package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"modern-social-media/internal/auth"
	"modern-social-media/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	firstNames = []string{
		"James", "John", "Robert", "Michael", "William", "David", "Richard", "Joseph", "Thomas", "Charles",
		"Mary", "Patricia", "Jennifer", "Linda", "Elizabeth", "Barbara", "Susan", "Jessica", "Sarah", "Karen",
		"Ahmed", "Mohamed", "Ali", "Fatima", "Aisha", "Omar", "Hassan", "Ibrahim", "Youssef", "Mariam",
	}
	lastNames = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
		"Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin",
		"Khan", "Ali", "Hussain", "Ahmed", "Rahman", "Malik", "Iqbal", "Shah", "Mahmood", "Hasan",
	}
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

	// 1. Clear existing users
	log.Println("Clearing existing users...")
	if err := db.Exec("TRUNCATE TABLE users CASCADE").Error; err != nil {
		log.Println("Truncate failed, trying Delete:", err)
		db.Where("1 = 1").Delete(&models.User{})
	}

	hashedPassword, _ := auth.HashPassword("password123")

	// 2. Prepare Avatar Pool (10 images)
	avatarDir := "uploads/avatars/random"
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		log.Fatal(err)
	}

	var avatarPaths []string
	log.Println("Preparing avatar pool (10 images)...")
	for i := 1; i <= 10; i++ {
		filename := fmt.Sprintf("pool_avatar_%d.png", i)
		localPath := filepath.Join(avatarDir, filename)
		webPath := fmt.Sprintf("/uploads/avatars/random/%s", filename)

		// Download if not exists
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			
			// Use different seed for each avatar in pool
			// We can just use the index 'i' as seed to be deterministic or consistent
			seed := fmt.Sprintf("pool_seed_%d", i)
			url := fmt.Sprintf("https://robohash.org/%s?set=set3&bgset=bg1", seed)
			
			if err := downloadAvatar(url, localPath); err != nil {
				log.Printf("Failed to download avatar %d: %v", i, err)
				continue
			}
			fmt.Printf("Downloaded pool avatar %d\n", i)
		} else {
			fmt.Printf("Using existing pool avatar %d\n", i)
		}
		
		avatarPaths = append(avatarPaths, webPath)
	}

	if len(avatarPaths) == 0 {
		log.Fatal("No avatars could be prepared")
	}

	// 3. Create Users
	log.Println("Creating users...")
	var users []models.User
	for i := 0; i < 50; i++ {
		first := firstNames[rand.Intn(len(firstNames))]
		last := lastNames[rand.Intn(len(lastNames))]
		suffix := fmt.Sprintf("%06d", rand.Intn(1000000))

		// Randomly pick one from the pool
		avatar := avatarPaths[rand.Intn(len(avatarPaths))]

		user := models.User{
			Username:     fmt.Sprintf("%s%s_%s", first, last, suffix),
			Email:        strings.ToLower(fmt.Sprintf("%s.%s.%s@example.com", first, last, suffix)),
			Password:     hashedPassword,
			FirstName:    first,
			LastName:     last,
			Bio:          fmt.Sprintf("Hello, I am %s!", first),
			AvatarURL:    avatar,
			IsActive:     true,
			IsVerified:   true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
			continue
		}
		users = append(users, user)
		fmt.Printf("Created user: %s (Avatar: %s)\n", user.Username, filepath.Base(avatar))
	}

	// 4. Create Friendships
	log.Println("Creating friendships...")
	successCount := 0
	errorCount := 0
	for _, user := range users {
		// Log checking ID validity
		if user.ID == "" {
			log.Printf("User %s has empty ID, skipping", user.Username)
			continue
		}

		numFriends := rand.Intn(10) + 1
		for j := 0; j < numFriends; j++ {
			friend := users[rand.Intn(len(users))]
			if friend.ID == user.ID {
				continue
			}

			// Force create instead of FirstOrCreate to verify insertion first, catching duplicate errors properly
			follow1 := models.Follow{
				FollowerID:  user.ID,
				FollowingID: friend.ID,
			}
			// Use FirstOrCreate to allow re-running seeder safely, but log error
			if err := db.FirstOrCreate(&follow1, models.Follow{FollowerID: user.ID, FollowingID: friend.ID}).Error; err != nil {
				log.Printf("Failed to create follow %s -> %s: %v", user.Username, friend.Username, err)
				errorCount++
				continue
			}

			follow2 := models.Follow{
				FollowerID:  friend.ID,
				FollowingID: user.ID,
			}
			if err := db.FirstOrCreate(&follow2, models.Follow{FollowerID: friend.ID, FollowingID: user.ID}).Error; err != nil {
				log.Printf("Failed to create follow %s -> %s: %v", friend.Username, user.Username, err)
				errorCount++
				continue
			}
			successCount += 2
		}
	}
	log.Printf("Friendship creation finished. Success: %d, Errors: %d", successCount, errorCount)

	fmt.Println("Seeding completed successfully")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func downloadAvatar(url, destPath string) error {
	resp, err := http.Get(url)
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
