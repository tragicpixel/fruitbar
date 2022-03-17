// Package JWT provides implementations of a JWT repository.
package jwt

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
)

// JWTRepository represents a repository of JSON Web Tokens.
type JWTRepository struct{}

// NewJWTRepository creates a new JWT repository.
func NewJWTRepository() repository.Jwt {
	return &JWTRepository{}
}

func (r *JWTRepository) GenerateToken(j *models.JwtWrapper, u *models.User) (signedToken string, err error) {
	claims := &models.JwtClaim{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(j.ExpirationHours)).Unix(),
			Issuer:    j.Issuer,
		},
		UserID:   u.ID,
		UserName: u.Name,
		UserRole: u.Role,
	}
	logrus.Info(fmt.Sprintf("Generated token with user id %d and role %s", claims.UserID, claims.UserRole))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err = token.SignedString([]byte(j.SecretKey))
	if err != nil {
		return
	}
	return
}

func (r *JWTRepository) ValidateToken(j *models.JwtWrapper, signedToken string) (claims *models.JwtClaim, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&models.JwtClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.SecretKey), nil
		},
	)
	if err != nil {
		return
	}
	claims, ok := token.Claims.(*models.JwtClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("JWT is expired")
		return
	}
	return
}

// same for username?
func (r *JWTRepository) GetRole(j *models.JwtWrapper, signedToken string) (role string, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&models.JwtClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.SecretKey), nil
		},
	)
	if err != nil {
		return
	}
	claims, ok := token.Claims.(*models.JwtClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}
	return claims.UserRole, nil
}
