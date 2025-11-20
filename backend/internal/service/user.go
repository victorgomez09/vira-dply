package service

import (
	"github.com/victorgomez09/vira-dply/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser crea un nuevo usuario en la base de datos.
func (s *UserService) CreateUser(username, email, password string) (*model.User, error) {
	password, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Email:    email,
		Password: password,
	}
	result := s.db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

// GetUserByUsername obtiene un usuario por su nombre de usuario.
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	result := s.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetAllUsers obtiene todos los usuarios de la base de datos.
func (s *UserService) GetAllUsers() ([]model.User, error) {
	var users []model.User
	result := s.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (s *UserService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// VerifyPassword verifies if the given password matches the stored hash.
func (s *UserService) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
