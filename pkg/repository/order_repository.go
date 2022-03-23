package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
)

// Order provides an interface for performing operations on a repository of orders.
type Order interface {
	// Count returns the count of all the orders based on the supplied seek options.
	Count(seek *PageSeekOptions) (count int64, err error)
	// Fetch returns the orders in the repository matching the supplied seek options.
	Fetch(seekOptions *PageSeekOptions) ([]*models.Order, error)
	// Exists determines if an order with the supplied id exists.
	Exists(id uint) (bool, error)
	// GetByID returns the order with the supplied id, if it exists.
	GetByID(id uint) (*models.Order, error)
	// Create creates a new order and returns the ID of the newly created product.
	Create(u *models.Order) (orderId uint, itemIds []uint, err error)
	// Update updates an existing order in the repository. Returns the updated order.
	Update(u *models.Order, fields []string) (*models.Order, error)
	// Delete removes an order with the supplied id from the repository.
	Delete(id uint) error
}
