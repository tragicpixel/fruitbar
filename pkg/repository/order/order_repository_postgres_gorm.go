// Package order provides implementations of an Order repository.
package order

import (
	"errors"

	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	"gorm.io/gorm"
)

// PostgresOrderRepo represents an implementation of an Order repository using postgres.
type PostgresOrderRepo struct {
	DB *gorm.DB
}

// NewPostgresOrderRepo creates a new postgres fruit order repository.
func NewPostgresOrderRepo(db *gorm.DB) repository.Order {
	return &PostgresOrderRepo{
		DB: db,
	}
}

func (r *PostgresOrderRepo) Count(seek *repository.PageSeekOptions) (count int64, err error) {
	var result *gorm.DB
	switch seek.Direction {
	case repository.SeekDirectionBefore:
		result = r.DB.Model(&models.Order{}).Where("ID < ?", seek.StartId).Count(&count)
	case repository.SeekDirectionAfter:
		result = r.DB.Model(&models.Order{}).Where("ID > ?", seek.StartId).Count(&count)
	case repository.SeekDirectionNone:
		result = r.DB.Model(&models.Order{}).Count(&count)
	default:
		return -1, errors.New("invalid seek direction")
	}
	if result.Error != nil {
		return -1, result.Error
	}
	return count, nil
}

func (r *PostgresOrderRepo) Fetch(seek *repository.PageSeekOptions) (orders []*models.Order, err error) {
	var result *gorm.DB
	if seek.Direction == repository.SeekDirectionBefore {
		result = r.DB.Limit(int(seek.RecordLimit)).Where("ID < ?", seek.StartId).Find(&orders)
	} else if seek.Direction == repository.SeekDirectionAfter {
		result = r.DB.Limit(int(seek.RecordLimit)).Where("ID > ?", seek.StartId).Find(&orders)
	} else if seek.Direction == repository.SeekDirectionNone {
		result = r.DB.Limit(int(seek.RecordLimit)).Find(&orders)
	} else {
		return nil, errors.New("invalid seek direction")
	}

	if result.Error != nil {
		return nil, result.Error
	}
	return orders, nil
}

func (r *PostgresOrderRepo) Exists(id uint) (exists bool, err error) {
	result := r.DB.Model(models.Order{}).Select("COUNT(*) > 0").Where("ID = ?", id).Find(&exists)
	if result.Error != nil {
		return false, result.Error
	}
	return exists, nil
}

func (r *PostgresOrderRepo) GetByID(id uint) (*models.Order, error) {
	var o models.Order
	result := r.DB.First(&o, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &o, nil
}

func (r *PostgresOrderRepo) Create(o *models.Order) (orderId uint, itemIds []uint, err error) {
	result := r.DB.Create(&o)
	if result.Error != nil {
		return 0, []uint{}, result.Error
	}
	for _, item := range o.Items {
		itemIds = append(itemIds, item.ID)
	}
	return o.ID, itemIds, nil
}

func (r *PostgresOrderRepo) Update(o *models.Order, fields []string) (update *models.Order, err error) {
	_, err = r.GetByID(o.ID)
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
	update, err = r.GetByID(o.ID)
	if err != nil {
		return nil, err
	}
	return update, nil
}

func (r *PostgresOrderRepo) Delete(id uint) error {
	// swap between these two based on some flag, set the flag in the deployment, so you can have different options for dev/test/prod builds
	//result := r.DB.Delete(&models.Order{}, id) // soft delete
	result := r.DB.Unscoped().Delete(&models.Order{}, id) // hard delete
	if result.Error != nil {
		return result.Error
	}
	return result.Error
}
