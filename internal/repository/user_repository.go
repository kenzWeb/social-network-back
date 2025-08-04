package repository

import (
	"modern-social-media/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}


func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}


func (r *UserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}


func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}


func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}


func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}


func (r *UserRepository) DeleteUser(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, id).Error
}


func (r *UserRepository) GetAllUsers(limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}


func (r *UserRepository) GetUserWithPosts(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Posts").Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}


func (r *UserRepository) SearchUsers(query string, limit, offset int) ([]models.User, error) {
	var users []models.User
	searchQuery := "%" + query + "%"
	err := r.db.Where("username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", 
		searchQuery, searchQuery, searchQuery).
		Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}
