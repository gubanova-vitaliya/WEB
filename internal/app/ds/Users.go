package ds

import (
	"time"

	"WEB/internal/app/role"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User - новая структура пользователя с UUID
type User struct {
	UUID      uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"uuid"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Login     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"login"`
	Email     string         `gorm:"type:varchar(100);uniqueIndex" json:"email"`
	Role      role.Role      `gorm:"type:varchar(20);not null;default:'buyer'" json:"role"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName задает имя таблицы
func (User) TableName() string {
	return "users"
}

// Users - старая структура (оставляем для обратной совместимости)
type Users struct {
	ID          uint   `gorm:"primary_key" json:"id"`
	Login       string `gorm:"type:varchar(25);unique;not null" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}

// TableName задает имя таблицы для старой структуры
func (Users) TableName() string {
	return "old_users"
}
