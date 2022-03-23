// Package product provides implementations of a Product repository.
package product

import (
	"errors"

	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	"gorm.io/gorm"
)

// PostgresProductRepo represents an implementation of a Product repository using postgres.
type PostgresProductRepo struct {
	DB *gorm.DB
}

// NewPostgresProductRepo creates a new postgres product repository.
func NewPostgresProductRepo(db *gorm.DB) repository.Product {
	return &PostgresProductRepo{
		DB: db,
	}
}

func (r *PostgresProductRepo) Count(seek *repository.PageSeekOptions) (count int64, err error) {
	var result *gorm.DB
	switch seek.Direction {
	case repository.SeekDirectionBefore:
		result = r.DB.Model(&models.Product{}).Where("ID < ?", seek.StartId).Count(&count)
	case repository.SeekDirectionAfter:
		result = r.DB.Model(&models.Product{}).Where("ID > ?", seek.StartId).Count(&count)
	case repository.SeekDirectionNone:
		result = r.DB.Model(&models.Product{}).Count(&count)
	default:
		return -1, errors.New("invalid seek direction")
	}
	if result.Error != nil {
		return -1, result.Error
	}
	return count, nil
}

func (r *PostgresProductRepo) Fetch(seek *repository.PageSeekOptions) (products []*models.Product, err error) {
	var result *gorm.DB
	switch seek.Direction {
	case repository.SeekDirectionBefore:
		result = r.DB.Limit(seek.RecordLimit).Where("ID < ?", seek.StartId).Find(&products)
	case repository.SeekDirectionAfter:
		result = r.DB.Limit(seek.RecordLimit).Where("ID > ?", seek.StartId).Find(&products)
	case repository.SeekDirectionNone:
		result = r.DB.Limit(seek.RecordLimit).Find(&products)
	default:
		return nil, errors.New("invalid seek direction")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

func (r *PostgresProductRepo) Exists(id uint) (bool, error) {
	var exists bool
	result := r.DB.Model(models.Product{}).Select("COUNT(*) > 0").Where("ID = ?", id).Find(&exists)
	if result.Error != nil {
		return false, result.Error
	}
	return exists, nil
}

func (r *PostgresProductRepo) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	result := r.DB.First(&product, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}

func (r *PostgresProductRepo) Create(p *models.Product) (uint, error) {
	result := r.DB.Create(&p)
	if result.Error != nil {
		return 0, result.Error
	}
	return p.ID, nil
}

func (r *PostgresProductRepo) Update(p *models.Product, fields []string) (*models.Product, error) {
	_, err := r.GetByID(p.ID)
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 { // Partial update
		result := r.DB.Model(p).Select(fields).Updates(p)
		if result.Error != nil {
			return nil, err
		}
	} else { // Full update
		result := r.DB.Model(p).Updates(p)
		if result.Error != nil {
			return nil, err
		}
	}
	updatedProduct, err := r.GetByID(p.ID) // rethink returning the updated product ... this doesn't return the fully updated product
	if err != nil {
		return nil, err
	}
	return updatedProduct, nil
}

func (r *PostgresProductRepo) Delete(id uint) error {
	// swap between these two based on some flag, set the flag in the deployment, so you can have different options for dev/test/prod builds
	//result := r.DB.Delete(&models.Product{}, id) // soft delete
	result := r.DB.Unscoped().Delete(&models.Product{}, id) // hard delete
	if result.Error != nil {
		return result.Error
	}
	return nil
}
