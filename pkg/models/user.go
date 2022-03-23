package models

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/tragicpixel/fruitbar/pkg/models/roles"
	"gorm.io/gorm"
)

// swagger:model user
// User holds information about a given user account.
type User struct {
	gorm.Model
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// PasswordFormatReqMsg returns a string containing all the formatting requirements for setting a password.
func PasswordFormatReqMsg() string {
	msg := fmt.Sprintf("Password must be between %d and %d characters long, ", passwordLengthMin, passwordLengthMax)
	msg += "contain at least one digit, "
	msg += "contain at least one of the following special characters: " + passwordValidSpecialChars
	return msg
}

// IsValid checks if a User object is valid.
func (u *User) IsValid() (bool, error) {
	_, err := u.validateName()
	if err != nil {
		return false, err
	}
	_, err = u.validatePassword()
	if err != nil {
		return false, err
	}
	_, err = roles.IsValid(u.Role)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ValidatePartialProductUpdate validates the supplied selected fields of the supplied product.
func (u *User) ValidatePartialUserUpdate(selectedFields []string) (bool, error) {
	var err error
	// TODO: Use code generation tools to extract the names in the json annotation
	for _, field := range selectedFields {
		switch field {
		case "name":
			_, err = u.validateName()
		case "password":
			_, err = u.validatePassword()
		case "role":
			_, err = roles.IsValid(u.Role)
		}
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

const (
	nameLengthMin = 1
	nameLengthMax = 100

	passwordLengthMin         = 8
	passwordLengthMax         = 100
	passwordMinDigitChars     = 1
	passwordMinLowercaseChars = 1
	passwordMinUppercaseChars = 1
	passwordMinSpecialChars   = 1
	passwordValidSpecialChars = "!@#$%^&*()-_=[];',./?~`"
)

// validateName checks if a user's current name is valid.
func (u *User) validateName() (bool, error) {
	length := len(u.Name)
	if length > nameLengthMax {
		return false, fmt.Errorf("name must be less than %d characters long", nameLengthMax)
	} else if length < nameLengthMin {
		return false, fmt.Errorf("name must be at least %d characters long", nameLengthMin)
	} else {
		return true, nil
	}
}

// validatePassword checks if a user's currently set password (plain text) is valid.
func (u *User) validatePassword() (bool, error) {
	length := len(u.Password)
	if length > passwordLengthMax {
		return false, fmt.Errorf("password must be less than %d characters long", passwordLengthMax)
	} else if length < passwordLengthMin {
		return false, fmt.Errorf("password must be at least %d characters long", passwordLengthMin)
	}

	containsDigit := false
	for _, char := range u.Password {
		if unicode.IsDigit(char) {
			containsDigit = true
			break
		}
	}
	if !containsDigit {
		return false, errors.New("password must contain at least one digit")
	}

	if !strings.ContainsAny(u.Password, passwordValidSpecialChars) {
		return false, errors.New("password must contain at least one of the following special characters: " + passwordValidSpecialChars)
	}

	return true, nil
}
