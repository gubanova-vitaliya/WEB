package repository

import (
	"WEB/internal/app/ds"

	"github.com/sirupsen/logrus"
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

// GetCartCount для получения количества услуг в заявке (чатов в сообщении в моем случае)
func (r *Repository) GetCartCount() int64 {
	var calculationID uint
	var count int64
	creatorID := 1
	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

	err := r.db.Model(&ds.Calculation{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("id").First(&calculationID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.GasCalculation{}).Where("message_id = ?", calculationID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_gas:", err)
	}

	return count
}
