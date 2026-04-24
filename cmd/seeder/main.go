package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
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
	rand.Seed(time.Now().UnixNano())

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

	// 2. Prepare Avatar Pool (curated, human-like styles + local fallback)
	avatarDir := "uploads/avatars/random"
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		log.Fatal(err)
	}

	var avatarPaths []string
	log.Println("Preparing avatar pool (24 images)...")
	avatarSeeds := buildAvatarSeeds(firstNames, lastNames)
	poolSize := 24
	for i := 0; i < poolSize; i++ {
		seed := avatarSeeds[i%len(avatarSeeds)]
		filename := fmt.Sprintf("pool_avatar_%02d.svg", i+1)
		localPath := filepath.Join(avatarDir, filename)
		webPath := fmt.Sprintf("/uploads/avatars/random/%s", filename)

		// Download only if missing to keep startup fast on subsequent runs.
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			avatarURL := buildAvatarURL(seed, i)
			if err := downloadAvatar(avatarURL, localPath); err != nil {
				initials := initialsFromSeed(seed)
				if svgErr := writeFallbackAvatarSVG(localPath, initials, i); svgErr != nil {
					log.Printf("Failed to generate avatar %d: %v", i+1, svgErr)
					continue
				}
				log.Printf("Avatar %d downloaded failed (%v). Fallback SVG generated.", i+1, err)
			} else {
				fmt.Printf("Prepared pool avatar %d\n", i+1)
			}
		} else {
			fmt.Printf("Using existing pool avatar %d\n", i+1)
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
			Username:   fmt.Sprintf("%s%s_%s", first, last, suffix),
			Email:      strings.ToLower(fmt.Sprintf("%s.%s.%s@example.com", first, last, suffix)),
			Password:   hashedPassword,
			FirstName:  first,
			LastName:   last,
			Bio:        fmt.Sprintf("Hello, I am %s!", first),
			AvatarURL:  avatar,
			IsActive:   true,
			IsVerified: true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
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
	client := &http.Client{Timeout: 12 * time.Second}
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

func buildAvatarSeeds(first, last []string) []string {
	var seeds []string
	for i := 0; i < len(first) && i < len(last); i++ {
		seeds = append(seeds, strings.ToLower(first[i]+"_"+last[i]))
	}

	for i := range first {
		seeds = append(seeds, strings.ToLower(first[i]))
	}

	sort.Strings(seeds)
	if len(seeds) == 0 {
		return []string{"default_user"}
	}
	return seeds
}

func buildAvatarURL(seed string, index int) string {
	styles := []string{"personas", "adventurer-neutral", "micah", "lorelei-neutral"}
	style := styles[index%len(styles)]
	escapedSeed := url.QueryEscape(seed)

	// DiceBear gives cleaner profile-like avatars than robohash set3.
	return fmt.Sprintf(
		"https://api.dicebear.com/9.x/%s/svg?seed=%s&size=256&backgroundType=gradientLinear",
		style,
		escapedSeed,
	)
}

func initialsFromSeed(seed string) string {
	parts := strings.FieldsFunc(seed, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	})
	if len(parts) == 0 {
		return "U"
	}

	first := strings.ToUpper(string(parts[0][0]))
	if len(parts) == 1 {
		return first
	}
	second := strings.ToUpper(string(parts[1][0]))
	return first + second
}

func writeFallbackAvatarSVG(path, initials string, index int) error {
	bgColors := []string{"#1D4ED8", "#047857", "#B45309", "#7C3AED", "#BE185D", "#0F766E"}
	bg := bgColors[index%len(bgColors)]

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256" role="img" aria-label="avatar">
  <rect width="256" height="256" rx="64" fill="%s"/>
  <text x="50%%" y="54%%" text-anchor="middle" dominant-baseline="middle" fill="#FFFFFF" font-family="Arial, Helvetica, sans-serif" font-size="96" font-weight="700">%s</text>
</svg>
`, bg, initials)

	return os.WriteFile(path, []byte(svg), 0644)
}
