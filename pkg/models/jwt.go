package models

import jwt "github.com/dgrijalva/jwt-go"

// JwtWrapper holds vital information about a given JSON web token for the fruitbar application.
type JwtWrapper struct {
	SecretKey       string
	Issuer          string
	ExpirationHours int64
}

// JwtClaim holds the standard JWT claim in addition to any other claims made that will need to be verified by the fruitbar application.
type JwtClaim struct {
	jwt.StandardClaims
	UserID   uint
	UserName string
	UserRole string
}
