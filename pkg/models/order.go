package models

import (
	"errors"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// CreditCardInfo holds the credit card information needed to run a credit card.
type CreditCardInfo struct {
	// Credit card number.
	Number string `json:"number"`
	// Name on the card.
	CardholderName string `json:"cardholdername"`
	// Expiration date. (mm/yy format)
	ExpirationDate string `json:"expirationdate"`
	// Zipcode on the card. (5 or 10 digits)
	Zipcode string `json:"zipcode"`
	// CVV on the card. (3 or 4 digits)
	Cvv string `json:"cvv"`
}

// PaymentInfo holds the payment information about a given order.
type PaymentInfo struct {
	// Whether the order paid in cash. (if false, this means it is paid with a credit card)
	Cash bool `json:"cash"`
	// Credit card info for this order.
	CardInfo CreditCardInfo `json:"cardinfo" gorm:"embedded"`
}

// swagger:model order
// Order holds all the information in an order of fruit.
type Order struct {
	gorm.Model
	// User ID of the user who owns the order.
	OwnerID uint `json:"ownerid"`
	// All of the items present in the order.
	Items []*Item `json:"items"`
	// Payment information for this order.
	PaymentInfo PaymentInfo `json:"paymentinfo" gorm:"embedded"`
	// Tax rate for this order.
	TaxRate float64 `json:"taxrate"`
	// Subtotal of the order. (before tax+tip)
	Subtotal float64 `json:"subtotal"`
	// Tax on the order.
	Tax float64 `json:"tax"`
	// Total cost of the order.
	Total float64 `json:"total"`
}

// ValidateCreditCardExpirationDate determines whether a credit card's expiration date is valid. (4 digit mm/yy string)
// TODO: write test functions for these -- how does go handle the / do i need to escape it? \/
func ValidateCreditCardExpirationDate(expDate string) error {
	match, _ := regexp.MatchString("[01][0123456789]/[0123456789]{2}", expDate)
	if !match {
		return errors.New("expiration date is invalid. Must be in MM/YY format and must be a valid date")
	}
	return nil
}

// ValidateCreditCardNumber determines whether a supplied credit card number is valid. (16 digit string)
func ValidateCreditCardNumber(cardNumber string) error {
	match, _ := regexp.MatchString("([0123456789]\\s*){16}", cardNumber)
	if !match {
		return errors.New("card number is invalid. Must be a 16 digit string, whitespace ignored")
	}
	return nil
}

// ValidateCreditCardCVV determines whether a supplied CVV is valid. (3-4 digit string)
func ValidateCreditCardCVV(cvv string) error {
	match, _ := regexp.MatchString("(([0123456789]\\s*){3}|([0123456789]\\s*){4})", cvv)
	if !match {
		return errors.New("the CVV is invalid. Must be a 3 or 4 digit string, whitespace ignored")
	}
	return nil
}

// ValidateZipcode determines whether a supplied zipcode is valid. (5 or 10 digit string)
func ValidateZipcode(zipcode string) error {
	match, _ := regexp.MatchString("([0123456789]){5}", zipcode)
	if !match {
		return errors.New("zipcode is invalid. Must be a 5 digit string")
	}
	return nil
}

// ValidateCreditCardInfo determines whether all of the supplied credit card information is valid.
func ValidateCreditCardInfo(cardInfo CreditCardInfo) error {
	expDateError := ValidateCreditCardExpirationDate(cardInfo.ExpirationDate)
	numberError := ValidateCreditCardNumber(cardInfo.Number)
	cvvError := ValidateCreditCardCVV(cardInfo.Cvv)
	zipcodeError := ValidateZipcode(cardInfo.Zipcode)
	if expDateError == nil && numberError == nil && cvvError == nil && zipcodeError == nil {
		return nil
	} else {
		errorMsgs := []string{expDateError.Error(), numberError.Error(), cvvError.Error(), zipcodeError.Error()}
		msg := strings.Join(errorMsgs, ", ")
		return errors.New(msg)
	}
}

// ValidateOrderPaymentInfo determines whether all of the supplied payment information for an order is valid.
func ValidateOrderPaymentInfo(info PaymentInfo) error {
	if info.Cash {
		return nil
	} else {
		return ValidateCreditCardInfo(info.CardInfo)
	}
}

// ValidateOrderId validates whether the supplied order's ID is valid. (if it exists)
func ValidateOrderId(order *Order) error {
	if order.ID == 0 { // 0 considered empty by go for an int
		return errors.New("ID cannot be null")
	}
	return nil
}

// ValidateOrder validates whether the supplied order is valid. (totals, payment info, and id need to be valid)
func ValidateOrder(order *Order) error {
	paymentInfoError := ValidateOrderPaymentInfo(order.PaymentInfo)
	idError := ValidateOrderId(order)
	totalIsValid := order.validateTotal()
	if paymentInfoError == nil && idError == nil && totalIsValid {
		return nil
	} else {
		errorMsgs := []string{paymentInfoError.Error(), idError.Error()}
		msg := strings.Join(errorMsgs, ", ")
		return errors.New(msg)
	}
}

// ValidateNewOrder validates whether the supplied new fruit order (freshly created) is valid. (payment info needs to be valid)
func ValidateNewOrder(order *Order) error {
	// Need also to validate that subtotal, tax, and total are empty??
	subtotalIsValid, taxIsValid, totalIsValid := true, true, true
	var subtotalError, taxError, totalError error
	if order.Subtotal != 0.0 {
		subtotalIsValid = false
		subtotalError = errors.New("subtotal must be empty")
	}
	if order.Tax != 0.0 {
		taxIsValid = false
		taxError = errors.New("tax must be empty")
	}
	if order.Total != 0.0 {
		totalIsValid = false
		totalError = errors.New("total must be empty")
	}
	paymentInfoError := ValidateOrderPaymentInfo(order.PaymentInfo)
	if subtotalIsValid && taxIsValid && totalIsValid && paymentInfoError == nil {
		return nil
	} else {
		errorMsgs := []string{subtotalError.Error(), taxError.Error(), totalError.Error(), paymentInfoError.Error()}
		msg := strings.Join(errorMsgs, ", ")
		return errors.New(msg)
	}
}

func ValidateOrderUpdate(order *Order, selectedFields []string) error {
	var err error
	valid := true
	for _, field := range selectedFields {
		switch field {
		case "ownerid":
			err = ValidateOrderId(order)
		case "paymentinfo":
			err = ValidateOrderPaymentInfo(order.PaymentInfo)
		case "items":
		case "taxrate":
		}
		if err != nil || !valid {
			return err
		}
	}
	return nil
}

func (o *Order) validateTax() bool {
	return (o.Tax == o.Subtotal*o.TaxRate)
}

func (o *Order) validateTotal() bool {
	return (o.Total == o.Subtotal+o.Tax)
}
