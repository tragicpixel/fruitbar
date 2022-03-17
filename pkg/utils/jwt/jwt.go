package jwt

import (
	"errors"
	"net/http"
	"strings"

	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
)

const (
	SECRET_KEY           = "verysecretkey"
	ISSUER               = "fruitbar"
	JWT_EXPIRATION_HOURS = 24
)

func GetSecretAuthToken() models.JwtWrapper {
	return models.JwtWrapper{
		SecretKey: SECRET_KEY,
		Issuer:    ISSUER,
	}
}

func GetTokenFromAuthHeader(auth string) (string, error) {
	token := strings.Split(auth, "Bearer ")
	if len(token) != 2 {
		return "", errors.New("incorrect format of authorization token")
	} else {
		auth = strings.TrimSpace(token[1])
		return auth, nil
	}
}

func GetTokenClaims(r *http.Request, jwtRepo repository.Jwt) (tokenClaims *models.JwtClaim, err error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return tokenClaims, errors.New("no authorization header provided")
	}
	authToken, err := GetTokenFromAuthHeader(auth)
	if err != nil {
		return tokenClaims, err
	}
	authReal := GetSecretAuthToken()
	tokenClaims, err = jwtRepo.ValidateToken(&authReal, authToken)
	if err != nil {
		return tokenClaims, errors.New("secretKey and/or Issuer wrong")
	}
	return tokenClaims, nil
}
