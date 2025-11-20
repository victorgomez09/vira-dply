package service

import (
	"github.com/victorgomez09/vira-dply/internal/model"
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
	user := &model.User{
		Username: username,
		Email:    email,
		Password: password, // En producción, asegúrate de hashear la contraseña
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
