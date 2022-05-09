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

// PasswordFmtReqMsg returns an array of strings containing all the formatting requirements for setting a password, where each item is a requirement.
func PasswordFmtReqMsg() []string {
	var msg []string
	msg = append(msg,
		fmt.Sprintf("Password must be between %d and %d characters long", passwordLengthMin, passwordLengthMax),
		"Password must contain at least one digit",
		"Password must contain at least one of the following special characters: "+passwordValidSpecialChars,
	)
	return msg
}

// IsValid checks if a User object is valid.
func (u *User) IsValid() error {
	err := u.validateName()
	if err != nil {
		return err
	}
	err = u.validatePassword()
	if err != nil {
		return err
	}
	err = roles.IsValid(u.Role)
	if err != nil {
		return err
	}
	return nil
}

// ValidatePartialProductUpdate validates the supplied selected fields of the supplied product.
func (u *User) ValidatePartialUserUpdate(selectedFields []string) error {
	var err error
	// TODO: Use code generation tools to extract the names in the json annotation
	for _, field := range selectedFields {
		switch field {
		case "name":
			err = u.validateName()
		case "password":
			err = u.validatePassword()
		case "role":
			err = roles.IsValid(u.Role)
		}
		if err != nil {
			return err
		}
	}
	return nil
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
func (u *User) validateName() error {
	length := len(u.Name)
	if length > nameLengthMax {
		return fmt.Errorf("name must be less than %d characters long", nameLengthMax)
	} else if length < nameLengthMin {
		return fmt.Errorf("name must be at least %d characters long", nameLengthMin)
	} else {
		return nil
	}
}

// validatePassword checks if a user's currently set password (plain text) is valid.
func (u *User) validatePassword() error {
	length := len(u.Password)
	if length > passwordLengthMax {
		return fmt.Errorf("password must be less than %d characters long", passwordLengthMax)
	} else if length < passwordLengthMin {
		return fmt.Errorf("password must be at least %d characters long", passwordLengthMin)
	}

	containsDigit := false
	for _, char := range u.Password {
		if unicode.IsDigit(char) {
			containsDigit = true
			break
		}
	}
	if !containsDigit {
		return errors.New("password must contain at least one digit")
	}

	if !strings.ContainsAny(u.Password, passwordValidSpecialChars) {
		return errors.New("password must contain at least one of the following special characters: " + passwordValidSpecialChars)
	}

	return nil
}
