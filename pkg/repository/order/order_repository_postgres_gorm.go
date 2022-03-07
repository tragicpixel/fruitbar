// Package order provides implementations of a FruitOrder repository.
package order

import (
	"errors"

	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	"github.com/tragicpixel/fruitbar/pkg/utils"
	"gorm.io/gorm"
)

// PostgresOrderRepo represents an implementation of a FruitOrder repository using postgres.
type PostgresOrderRepo struct {
	DB *gorm.DB
}

// NewPostgresOrderRepo creates a new postgres fruit order repository.
func NewPostgresOrderRepo(db *gorm.DB) repository.Order {
	return &PostgresOrderRepo{
		DB: db,
	}
}

func (r *PostgresOrderRepo) Fetch(pageSeekOptions utils.PageSeekOptions) ([]*models.Order, error) {
	var orders []*models.Order
	var result *gorm.DB
	if pageSeekOptions.Direction == utils.SeekDirectionBefore {
		result = r.DB.Limit(int(pageSeekOptions.RecordLimit)).Where("ID < ?", pageSeekOptions.StartId).Find(&orders)
	} else if pageSeekOptions.Direction == utils.SeekDirectionAfter {
		result = r.DB.Limit(int(pageSeekOptions.RecordLimit)).Where("ID > ?", pageSeekOptions.StartId).Find(&orders)
	} else if pageSeekOptions.Direction == utils.SeekDirectionNone {
		result = r.DB.Limit(int(pageSeekOptions.RecordLimit)).Find(&orders)
	} else {
		return nil, errors.New("invalid seek direction")
	}

	if result.Error != nil {
		return nil, result.Error
	}
	return orders, nil
}

func (r *PostgresOrderRepo) Exists(id int) (bool, error) {
	var exists bool
	result := r.DB.Model(models.Order{}).Select("COUNT(*) > 0").Where("ID = ?", id).Find(&exists)
	if result.Error != nil {
		return false, result.Error
	}
	return exists, nil
}

func (r *PostgresOrderRepo) GetByID(id int64) (*models.Order, error) {
	var order models.Order
	result := r.DB.First(&order, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &order, nil
}

func (r *PostgresOrderRepo) Create(o *models.Order) (int64, error) {
	result := r.DB.Create(&o)
	if result.Error != nil {
		return -1, result.Error
	}
	return int64(o.ID), nil
}

func (r *PostgresOrderRepo) Update(o *models.Order, fields []string) (*models.Order, error) {
	_, err := r.GetByID(int64(o.ID))
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 { // Partial update
		result := r.DB.Model(o).Select(fields).Updates(o)
		if result.Error != nil {
			return nil, err
		}
	} else { // Full update
		result := r.DB.Model(o).Updates(o)
		if result.Error != nil {
			return nil, err
		}
	}
	updatedOrder, err := r.GetByID(int64(o.ID))
	if err != nil {
		return nil, err
	}
	return updatedOrder, nil
}

func (r *PostgresOrderRepo) Delete(id int64) (bool, error) {
	// swap between these two based on some flag, set the flag in the deployment, so you can have different options for dev/test/prod builds
	//result := r.DB.Delete(&models.FruitOrder{}, id) // soft delete
	result := r.DB.Unscoped().Delete(&models.Order{}, id) // hard delete
	if result.Error != nil {
		return false, result.Error
	}
	return true, result.Error
}
