// Package order provides implementations of a FruitOrder repository.
package order

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
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

func (r *PostgresOrderRepo) Fetch(num int64) ([]*models.FruitOrder, error) {
	var orders []*models.FruitOrder
	result := r.DB.Limit(int(num)).Find(&orders)
	if result.Error != nil {
		return nil, result.Error
	}
	return orders, nil
}

func (r *PostgresOrderRepo) GetByID(id int64) (*models.FruitOrder, error) {
	var order models.FruitOrder
	result := r.DB.First(&order, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &order, nil
}

func (r *PostgresOrderRepo) Create(o *models.FruitOrder) (int64, error) {
	result := r.DB.Create(&o)
	if result.Error != nil {
		return -1, result.Error
	}
	return int64(o.ID), nil
}

func (r *PostgresOrderRepo) Update(o *models.FruitOrder) (*models.FruitOrder, error) {
	existingOrder, err := r.GetByID(int64(o.ID))
	if err != nil {
		return nil, err
	}
	existingOrder = o
	result := r.DB.Save(&existingOrder)
	if result.Error != nil {
		return nil, err
	}
	return existingOrder, nil
}

func (r *PostgresOrderRepo) Delete(id int64) (bool, error) {
	// swap between these two based on some flag, set the flag in the deployment, so you can have different options for dev/test/prod builds
	//result := r.DB.Delete(&models.FruitOrder{}, id) // soft delete
	result := r.DB.Unscoped().Delete(&models.FruitOrder{}, id) // hard delete
	if result.Error != nil {
		return false, result.Error
	}
	return true, result.Error
}
