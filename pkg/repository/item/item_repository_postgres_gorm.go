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

func (r *PostgresItemRepo) Fetch(pageSeekOptions utils.PageSeekOptions) ([]*models.Item, error) {
	var items []*models.Item
	var result *gorm.DB
	if pageSeekOptions.Direction == utils.SEEK_DIRECTION_BEFORE {
		result = r.DB.Limit(pageSeekOptions.RecordLimit).Where("ID < ?", pageSeekOptions.StartId).Find(&items)
	} else if pageSeekOptions.Direction == utils.SEEK_DIRECTION_AFTER {
		result = r.DB.Limit(pageSeekOptions.RecordLimit).Where("ID > ?", pageSeekOptions.StartId).Find(&items)
	} else if pageSeekOptions.Direction == utils.SEEK_DIRECTION_NONE {
		result = r.DB.Limit(pageSeekOptions.RecordLimit).Find(&items)
	} else {
		return nil, errors.New("invalid seek direction")
	}

	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (r *PostgresItemRepo) Exists(id int) (bool, error) {
	var exists bool
	result := r.DB.Model(models.Item{}).Select("COUNT(*) > 0").Where("ID = ?", id).Find(&exists)
	if result.Error != nil {
		return false, result.Error
	}
	return exists, nil
}

func (r *PostgresItemRepo) GetByID(id int) (*models.Item, error) {
	var item models.Item
	result := r.DB.First(&item, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}

func (r *PostgresItemRepo) GetByOrderID(id int) ([]*models.Item, error) {
	var items []*models.Item
	result := r.DB.Where(&models.Item{OrderID: id}).Find(&items)
	if result.Error != nil { // TODO: && result.Error != gorm.ErrRecordNotFound ??? test this
		return nil, result.Error
	} else {
		return items, nil
	}
}

func (r *PostgresItemRepo) Create(i *models.Item) (int64, error) {
	result := r.DB.Create(&i)
	if result.Error != nil {
		return -1, result.Error
	}
	return int64(i.ID), nil
}

func (r *PostgresItemRepo) Update(i *models.Item, fields []string) (*models.Item, error) {
	_, err := r.GetByID(int(i.ID))
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
	updated, err := r.GetByID(int(i.ID)) // TODO: fix -- this doesnt return the updated item
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *PostgresItemRepo) Delete(id int64) (bool, error) {
	// swap between these two based on some flag, set the flag in the deployment, so you can have different options for dev/test/prod builds
	//result := r.DB.Delete(&models.Item{}, id) // soft delete
	result := r.DB.Unscoped().Delete(&models.Item{}, id) // hard delete
	if result.Error != nil {
		return false, result.Error
	}
	return true, result.Error
}
