package repository

import (
	"WEB/internal/app/ds"
)

func (r *Repository) GetAllGases() ([]ds.Gas, error) {
	var gas []ds.Gas
	err := r.db.Find(&gas).Error
	if err != nil {
		return nil, err
	}
	return gas, nil
}

func (r *Repository) GetGasByID(id int) (*ds.Gas, error) {
	var gas ds.Gas
	err := r.db.First(&gas, id).Error
	if err != nil {
		return nil, err
	}
	return &gas, nil
}

func (r *Repository) SearchGasesByTitle(title string) ([]ds.Gas, error) {
	var gas []ds.Gas
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&gas).Error
	if err != nil {
		return nil, err
	}
	return gas, nil
}

// УДАЛИТЬ весь метод GetCartCount отсюда - он уже есть в calculation.go
