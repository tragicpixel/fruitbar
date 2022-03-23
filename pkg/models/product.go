package models

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"gorm.io/gorm"
)

// swagger:model product
// Product represents a product that can be purchased in individual units and added as items to an order.
type Product struct {
	gorm.Model
	// Name of the product.
	Name string `json:"name"`
	// The single rune used to represent the product. (if applicable) TODO: use rune type??
	Symbol string `json:"symbol"`
	// Price of the product, in dollars.
	Price float64 `json:"price"`
	// Number of the product currently in stock.
	NumInStock int `json:"numInStock"`
}

func (p *Product) IsValid() (bool, error) {
	errMsgPrefix := "failed to validate product: "
	_, err := p.nameIsValid()
	if err != nil {
		return false, errors.New(errMsgPrefix + err.Error())
	}
	_, err = p.symbolIsValid()
	if err != nil {
		return false, errors.New(errMsgPrefix + err.Error())
	}
	_, err = p.priceIsValid()
	if err != nil {
		return false, errors.New(errMsgPrefix + err.Error())
	}
	_, err = p.numInStockIsValid()
	if err != nil {
		return false, errors.New(errMsgPrefix + err.Error())
	}
	return true, nil
}

func (p *Product) PartialUpdateIsValid(selectedFields []string) (bool, error) {
	var err error
	// TODO: Use code generation tools to extract the names of the json annotations and use them here
	for _, field := range selectedFields {
		switch field {
		case "name":
			_, err = p.nameIsValid()
		case "symbol":
			_, err = p.symbolIsValid()
		case "price":
			_, err = p.priceIsValid()
		case "numInStock":
			_, err = p.numInStockIsValid()
		default:
			err = fmt.Errorf("field name is invalid: %s", field)
		}
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (p *Product) nameIsValid() (bool, error) {
	if len(p.Name) < 1 {
		return false, errors.New("name must be at least 1 character")
	}
	return true, nil
}

func (p *Product) symbolIsValid() (bool, error) {
	errMsg := "symbol must be exactly 1 rune"
	length := utf8.RuneCountInString(p.Symbol)
	if length > 1 {
		return false, errors.New(errMsg)
	} else if length < 1 {
		return false, errors.New(errMsg)
	} else {
		return true, nil
	}
}

func (p *Product) priceIsValid() (bool, error) {
	if p.Price <= 0 {
		return false, fmt.Errorf("price must be greater than zero, got %.2f", p.Price)
	} else {
		return true, nil
	}
}

func (p *Product) numInStockIsValid() (bool, error) {
	if p.NumInStock < 0 {
		return false, fmt.Errorf("numInStock must be greater than or equal to zero, got %d", p.NumInStock)
	} else {
		return true, nil
	}
}
