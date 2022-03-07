package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils"
)

// Order provides an interface for performing operations on a repoistory of orders.
type Order interface {
	// Fetch find and returns an array of all orders in the repository. Returns nil on error.
	Fetch(pageSeekOptions utils.PageSeekOptions) ([]*models.Order, error)
	// Exists determines if an order with the supplied id exists.
	Exists(id int) (bool, error)
	// GetByID finds and returns an individual order with the supplied id. Returns nil on error.
	GetByID(num int64) (*models.Order, error)
	// Create creates a new order and places it in the repository. Returns the ID of the newly created order, -1 on error.
	Create(u *models.Order) (orderId uint, itemIds []uint, err error)
	// Update updates an existing order in the repository. Returns nil on error.
	Update(u *models.Order, fields []string) (*models.Order, error)
	// Delete removes an order with the supplied id from the repository. Returns true on success, false on error.
	Delete(id int64) (bool, error)
}
