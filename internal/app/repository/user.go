package repository

import (
	"WEB/internal/app/ds"
	"WEB/internal/app/role"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Register регистрирует нового пользователя
func (r *Repository) Register(user *ds.User) error {
	if user.UUID == uuid.Nil {
		user.UUID = uuid.New()
	}

	// Проверяем, не существует ли уже пользователь с таким логином
	var existingUser ds.User
	if err := r.db.Where("login = ?", user.Login).First(&existingUser).Error; err == nil {
		return errors.New("user with this login already exists")
	}

	// Проверяем email, если он предоставлен
	if user.Email != "" {
		if err := r.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			return errors.New("user with this email already exists")
		}
	}

	return r.db.Create(user).Error
}

// GenerateHashString создает хеш пароля с использованием bcrypt
func (r *Repository) GenerateHashString(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword проверяет соответствие пароля и хеша
func (r *Repository) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GetUserByLogin возвращает пользователя по логину
func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var user ds.User
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// AuthenticateUser аутентифицирует пользователя
func (r *Repository) AuthenticateUser(login, password string) (*ds.User, error) {
	user, err := r.GetUserByLogin(login)
	if err != nil {
		return nil, errors.New("invalid login or password")
	}

	// Проверяем пароль с использованием bcrypt
	err = r.VerifyPassword(user.Password, password)
	if err != nil {
		return nil, errors.New("invalid login or password")
	}

	return user, nil
}

// GetUserByUUID возвращает пользователя по UUID
func (r *Repository) GetUserByUUID(userUUID string) (*ds.User, error) {
	var user ds.User
	if err := r.db.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserRole обновляет роль пользователя (только для администраторов)
func (r *Repository) UpdateUserRole(userUUID string, newRole role.Role) error {
	return r.db.Model(&ds.User{}).Where("uuid = ?", userUUID).Update("role", newRole).Error
}

// GetAllUsers возвращает всех пользователей (для администраторов)
func (r *Repository) GetAllUsers() ([]ds.User, error) {
	var users []ds.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// DeleteUser мягко удаляет пользователя
func (r *Repository) DeleteUser(userUUID string) error {
	return r.db.Where("uuid = ?", userUUID).Delete(&ds.User{}).Error
}
