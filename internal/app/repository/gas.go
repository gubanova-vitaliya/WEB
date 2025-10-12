package repository

import (
	"WEB/internal/app/ds"
)

func (r *Repository) GetAllGases() ([]ds.Gas, error) {
	// тут мы пользуемся ORM
	var gas []ds.Gas
	err := r.db.Find(&gas).Error
	if err != nil {
		return nil, err
	}
	return gas, nil
}

func (r *Repository) SearchGasesByTitle(title string) ([]ds.Gas, error) {
	var gas []ds.Gas
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&gas).Error
	if err != nil {
		return nil, err
	}
	return gas, nil
}
