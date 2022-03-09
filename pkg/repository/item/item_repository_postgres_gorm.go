package item

import (
	"errors"

	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	"github.com/tragicpixel/fruitbar/pkg/utils"
	"gorm.io/gorm"
)

// PostgresProductRepo represents an implementation of a Product repository using postgres.
type PostgresItemRepo struct {
	DB *gorm.DB
}

// NewPostgresProductRepo creates a new postgres product repository.
func NewPostgresItemRepo(db *gorm.DB) repository.Item {
	return &PostgresItemRepo{
		DB: db,
	}
}

func (r *PostgresItemRepo) Count(seek *utils.PageSeekOptions) (count int64, err error) {
	var result *gorm.DB
	switch seek.Direction {
	case utils.SeekDirectionBefore:
		result = r.DB.Model(&models.Item{}).Where("ID < ?", seek.StartId).Count(&count)
	case utils.SeekDirectionAfter:
		result = r.DB.Model(&models.Item{}).Where("ID > ?", seek.StartId).Count(&count)
	case utils.SeekDirectionNone:
		result = r.DB.Model(&models.Item{}).Count(&count)
	default:
		return -1, errors.New("invalid seek direction")
	}
	if result.Error != nil {
		return -1, result.Error
	}
	return count, nil
}

func (r *PostgresItemRepo) Fetch(seek *utils.PageSeekOptions) (items []*models.Item, err error) {
	var result *gorm.DB
	switch seek.Direction {
	case utils.SeekDirectionBefore:
		result = r.DB.Limit(seek.RecordLimit).Where("ID < ?", seek.StartId).Find(&items)
	case utils.SeekDirectionAfter:
		result = r.DB.Limit(seek.RecordLimit).Where("ID > ?", seek.StartId).Find(&items)
	case utils.SeekDirectionNone:
		result = r.DB.Limit(seek.RecordLimit).Find(&items)
	default:
		return nil, errors.New("invalid seek direction")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (r *PostgresItemRepo) Exists(id uint) (bool, error) {
	var exists bool
	result := r.DB.Model(models.Item{}).Select("COUNT(*) > 0").Where("ID = ?", id).Find(&exists)
	if result.Error != nil {
		return false, result.Error
	}
	return exists, nil
}

func (r *PostgresItemRepo) GetByID(id uint) (*models.Item, error) {
	var item models.Item
	result := r.DB.First(&item, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

func (r *PostgresItemRepo) GetByOrderID(id uint) ([]*models.Item, error) {
	var items []*models.Item
	result := r.DB.Where(&models.Item{OrderID: id}).Find(&items)
	if result.Error != nil { // TODO: && result.Error != gorm.ErrRecordNotFound ??? test this
		return nil, result.Error
	} else {
		return items, nil
	}
}

func (r *PostgresItemRepo) GetByProductID(id uint) ([]*models.Item, error) {
	var items []*models.Item
	result := r.DB.Where(&models.Item{ProductID: id}).Find(&items)
	if result.Error != nil { // TODO: && result.Error != gorm.ErrRecordNotFound ??? test this
		return nil, result.Error
	} else {
		return items, nil
	}
}

func (r *PostgresItemRepo) Create(i *models.Item) (uint, error) {
	result := r.DB.Create(&i)
	if result.Error != nil {
		return 0, result.Error
	}
	return i.ID, nil
}

func (r *PostgresItemRepo) Update(i *models.Item, fields []string) (*models.Item, error) {
	_, err := r.GetByID(i.ID)
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 { // Partial update
		result := r.DB.Model(i).Select(fields).Updates(i)
		if result.Error != nil {
			return nil, err
		}
	} else { // Full update
		result := r.DB.Model(i).Updates(i)
		if result.Error != nil {
			return nil, err
		}
	}
	updated, err := r.GetByID(i.ID) // TODO: fix -- this doesnt return the updated item
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *PostgresItemRepo) Delete(id uint) error {
	// swap between these two based on some flag, set the flag in the deployment, so you can have different options for dev/test/prod builds
	//result := r.DB.Delete(&models.Item{}, id) // soft delete
	result := r.DB.Unscoped().Delete(&models.Item{}, id) // hard delete
	if result.Error != nil {
		return result.Error
	}
	return result.Error
}
