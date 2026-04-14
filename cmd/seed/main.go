package main

import (
	"fmt"
	"log"
	"os"

	"war-room-backend/internal/config"
	"war-room-backend/internal/db"
	"war-room-backend/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	db.Connect(cfg)

	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	name := os.Getenv("ADMIN_NAME")

	if email == "" {
		email = "admin@warroom.com"
	}
	if password == "" {
		password = "admin123"
	}
	if name == "" {
		name = "Admin"
	}

	// Check if admin already exists
	var existing models.User
	err := db.DB.Where("email = ?", email).First(&existing).Error
	if err == nil {
		// Already exists — ensure role is admin
		if existing.Role != "admin" {
			db.DB.Model(&existing).Update("role", "admin")
			fmt.Printf("Updated existing user %s to admin role\n", email)
		} else {
			fmt.Printf("Admin user %s already exists\n", email)
		}
		return
	}
	if err != gorm.ErrRecordNotFound {
		log.Fatalf("Database error: %v", err)
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Could not hash password: %v", err)
	}

	admin := models.User{
		ID:       uuid.New().String(),
		Email:    email,
		Password: string(hashed),
		Name:     name,
		Role:     "admin",
	}

	if err := db.DB.Create(&admin).Error; err != nil {
		log.Fatalf("Could not create admin: %v", err)
	}

	fmt.Printf("Admin user created: %s / %s\n", email, password)
}
