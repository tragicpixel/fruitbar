package models

import (
	"errors"

	"gorm.io/gorm"
)

// swagger:model item
type Item struct {
	gorm.Model
	OrderID   uint `json:"orderid"`
	ProductID uint `json:"productid"`
	Quantity  int  `json:"quantity"`
}

func (i *Item) ValidateOrderID() (bool, error) {
	if i.OrderID <= 0 {
		return false, errors.New("orderid must be greater than zero")
	}
	return true, nil
}

func (i *Item) ValidateProductID() (bool, error) {
	if i.ProductID <= 0 {
		return false, errors.New("productid must be greater than zero")
	}
	return true, nil
}

func (i *Item) ValidateQuantity() (bool, error) {
	if i.Quantity <= 0 {
		return false, errors.New("quantity must be greater than zero")
	}
	return true, nil
}
