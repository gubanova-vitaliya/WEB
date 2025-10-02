package repository

import (
	"WEB/internal/app/ds"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // подключаемся к БД
	if err != nil {
		return nil, err
	}

	// Автомиграция для создания таблиц
	err = db.AutoMigrate(
		&ds.Gas{},
		&ds.Calculation{},
		&ds.CalculationGas{},
		&ds.Users{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	// Возвращаем объект Repository с подключенной базой данных
	return &Repository{
		db: db,
	}, nil
}
