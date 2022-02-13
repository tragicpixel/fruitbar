// Package user provides implementations of a user account repository.
package user

import (
	"context"

	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
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

func (r *PostgresUserRepo) Fetch(ctx context.Context, num int64) ([]*models.User, error) {
	var users []*models.User
	result := r.DB.Limit(int(num)).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *PostgresUserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
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

func (r *PostgresUserRepo) Create(ctx context.Context, u *models.User) (int64, error) {
	result := r.DB.Create(&u)
	if result.Error != nil {
		return -1, result.Error
	}
	return int64(u.ID), nil
}

func (r *PostgresUserRepo) Update(ctx context.Context, u *models.User) (*models.User, error) {
	existingUser, err := r.GetByID(ctx, int64(u.ID))
	if err != nil {
		return nil, err
	}
	existingUser = u
	result := r.DB.Save(&existingUser)
	if result.Error != nil {
		return nil, err
	}
	return existingUser, nil
}

func (r *PostgresUserRepo) Delete(ctx context.Context, id int64) (bool, error) {
	// swap between these two based on some flag, set the flag in the deployment, so you can have different options for dev/test/prod builds
	//result := r.DB.Delete(&models.User{}, id) // soft delete
	result := r.DB.Unscoped().Delete(&models.User{}, id) // hard delete
	if result.Error != nil {
		return false, result.Error
	}
	return true, result.Error
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
