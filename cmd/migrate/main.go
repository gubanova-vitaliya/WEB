package main

import (
	"WEB/internal/app/ds"
	"WEB/internal/app/dsn"
	"WEB/internal/app/role"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Включаем расширение UUID
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// Migrate the schema
	err = db.AutoMigrate(
		&ds.Gas{},
		&ds.Calculation{},
		&ds.GasCalculation{},
		&ds.Users{},
		&ds.User{}, // Добавляем новую таблицу пользователей
	)
	if err != nil {
		panic("cant migrate db")
	}

	// Создаем администратора по умолчанию
	createDefaultAdmin(db)
}

func createDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&ds.User{}).Where("role = ?", role.Admin).Count(&count)

	if count == 0 {
		adminUser := &ds.User{
			Name:     "Administrator",
			Login:    "admin",
			Email:    "admin@gaseproject.com",
			Role:     role.Admin,
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		}

		if err := db.Create(adminUser).Error; err != nil {
			println("Warning: Failed to create default admin user:", err.Error())
		} else {
			println("Default admin user created: login='admin', password='password'")
		}
	}
}
