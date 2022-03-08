package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils"
)

// Order provides an interface for performing operations on a repoistory of orders.
type Order interface {
	// Fetch returns a slice of orders based on the supplied seek options.
	Fetch(seekOptions *utils.PageSeekOptions) ([]*models.Order, error)
	// Exists determines if an order with the supplied id exists.
	Exists(id uint) (bool, error)
	// GetByID returns the order with the supplied id.
	GetByID(id uint) (*models.Order, error)
	// Create creates a new order and places it in the repository.
	// Returns the ID of the new order and the IDs of the new items created for the order.
	Create(u *models.Order) (orderId uint, itemIds []uint, err error)
	// Update updates an existing order in the repository. Returns the updated order.
	Update(u *models.Order, fields []string) (*models.Order, error)
	// Delete removes an order with the supplied id from the repository. Returns true on success.
	Delete(id uint) (bool, error)
}
