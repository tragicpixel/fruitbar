// Package user provides implementations of a user account repository.
package user

import (
	"errors"

	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	"github.com/tragicpixel/fruitbar/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TODO: Remove context

// PostgresUserRepo represents an implementation of a user account repository using postgres.
type PostgresUserRepo struct {
	DB *gorm.DB
}

// NewPostgresUserRepo creates a new postgres user account repository.
func NewPostgresUserRepo(db *gorm.DB) repository.User {
	return &PostgresUserRepo{
		DB: db,
	}
}

func (r *PostgresUserRepo) Count(seek *utils.PageSeekOptions) (count int64, err error) {
	var result *gorm.DB
	switch seek.Direction {
	case utils.SeekDirectionBefore:
		result = r.DB.Model(&models.User{}).Where("ID < ?", seek.StartId).Count(&count)
	case utils.SeekDirectionAfter:
		result = r.DB.Model(&models.User{}).Where("ID > ?", seek.StartId).Count(&count)
	case utils.SeekDirectionNone:
		result = r.DB.Model(&models.User{}).Count(&count)
	default:
		return -1, errors.New("invalid seek direction")
	}
	if result.Error != nil {
		return -1, result.Error
	}
	return count, nil
}

func (r *PostgresUserRepo) Fetch(seek *utils.PageSeekOptions) (users []*models.User, err error) {
	var result *gorm.DB
	switch seek.Direction {
	case utils.SeekDirectionBefore:
		result = r.DB.Limit(seek.RecordLimit).Where("ID < ?", seek.StartId).Find(&users)
	case utils.SeekDirectionAfter:
		result = r.DB.Limit(seek.RecordLimit).Where("ID > ?", seek.StartId).Find(&users)
	case utils.SeekDirectionNone:
		result = r.DB.Limit(seek.RecordLimit).Find(&users)
	default:
		return nil, errors.New("invalid seek direction")
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *PostgresUserRepo) Exists(id uint) (bool, error) {
	var exists bool
	result := r.DB.Model(models.User{}).Select("COUNT(*) > 0").Where("ID = ?", id).Find(&exists)
	if result.Error != nil {
		return false, result.Error
	}
	return exists, nil
}

func (r *PostgresUserRepo) GetByID(id uint) (*models.User, error) {
	var user models.User
	result := r.DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *PostgresUserRepo) GetByUsername(uname string) (*models.User, error) {
	var user models.User
	result := r.DB.Limit(1).Where("name = ?", uname).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *PostgresUserRepo) Create(u *models.User) (uint, error) {
	result := r.DB.Create(&u)
	if result.Error != nil {
		return 0, result.Error
	}
	return u.ID, nil
}

func (r *PostgresUserRepo) Update(u *models.User, fields []string) (*models.User, error) {
	_, err := r.GetByID(u.ID)
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 { // Partial update
		result := r.DB.Model(u).Select(fields).Updates(u)
		if result.Error != nil {
			return nil, err
		}
	} else { // Full update
		result := r.DB.Model(u).Updates(u)
		if result.Error != nil {
			return nil, err
		}
	}
	updated, err := r.GetByID(u.ID) // TODO: fix, doesn't return the updated product??
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *PostgresUserRepo) Delete(id uint) error {
	// swap between these two based on some flag, set the flag in the deployment, so you can have different options for dev/test/prod builds
	//result := r.DB.Delete(&models.User{}, id) // soft delete
	result := r.DB.Unscoped().Delete(&models.User{}, id) // hard delete
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *PostgresUserRepo) HashPassword(u *models.User, pass string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

func (r *PostgresUserRepo) CheckPassword(u *models.User, rawPass string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(rawPass))
	if err != nil {
		return err
	}
	return nil
}
