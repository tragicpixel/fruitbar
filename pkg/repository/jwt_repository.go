package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
)

// Jwt provides an interface for generating and validating JSON web tokens.
type Jwt interface {
	// GenerateToken returns a signed token based on the supplied JWT wrapper.
	GenerateToken(j *models.JwtWrapper) (signedToken string, err error)
	// ValidateToken returns a JWT claim based on the supplied JWT wrapper and signed token.
	ValidateToken(j *models.JwtWrapper, signedToken string) (claims *models.JwtClaim, err error)
}
