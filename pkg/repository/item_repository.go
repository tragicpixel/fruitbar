package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
)

// Order provides an interface for performing operations on a repoistory of orders.
type Item interface {
	// Count returns the count of all the records matching the supplied seek options.
	Count(seek *PageSeekOptions) (count int64, err error)
	// Fetch returns the records in the repository matching the supplied seek options.
	Fetch(pageSeekOptions *PageSeekOptions) ([]*models.Item, error)
	// Exists determines if an item with the supplied id exists.
	Exists(id uint) (bool, error)
	// GetByID returns the item with the supplied id, if it exists.
	GetByID(id uint) (*models.Item, error)
	// GetByOrderID returns an array of all the items in the repository with the supplied order id.
	GetByOrderID(id uint) ([]*models.Item, error)
	// GetByOrderID returns an array of all the items in the repository with the supplied product id.
	GetByProductID(id uint) ([]*models.Item, error)
	// Create creates a new record and returns its ID.
	Create(i *models.Item) (uint, error)
	// Update updates an existing product in the repository and returns the updated record.
	Update(i *models.Item, fields []string) (*models.Item, error)
	// Delete removes the record with the supplied id from the repository.
	Delete(id uint) error
}
