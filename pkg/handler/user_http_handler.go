package handler

import (
	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	jwtrepo "github.com/tragicpixel/fruitbar/pkg/repository/jwt"
	userrepo "github.com/tragicpixel/fruitbar/pkg/repository/user"
	"github.com/tragicpixel/fruitbar/pkg/utils"
	"gorm.io/gorm"

	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// TODO: How to properly store these???
const (
	SECRET_KEY = "verysecretkey"
	ISSUER     = "fruitbar"
	UI_URL     = "http://localhost:3000"
)

// User represents an http handler for performing operations on a repository of user accounts.
type User struct {
	repo    repository.User
	jwtRepo repository.Jwt
}

// NewUserHandler creates an http handler for performing operations on a repository of user accounts.
func NewUserHandler(db *driver.DB) *User {
	return &User{
		repo:    userrepo.NewPostgresUserRepo(db.Postgres), // this is where it is decided which implementation(/database type) of the User Repo we will use
		jwtRepo: jwtrepo.NewJWTRepository(),
	}
}

// RegisterNewUser creates a new user with the specified username and password, and returns a JSON message if there is an error, otherwise no content.
func (h *User) RegisterNewUser(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodPost, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Register API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}
	logrus.Info("Starting login process...")

	response := utils.JsonResponse{}
	user := models.User{}
	err := utils.DecodeJSONBody(w, r, &user, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if err != nil {
		var request *utils.MalformedRequestError
		if errors.As(err, &request) {
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: request.Status, Message: request.Message}}
		} else {
			msg := "failed to decode JSON body: " + err.Error()
			logrus.Error(msg)
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: msg}}
		}
	}

	existingUser, _ := h.repo.GetByUsername(user.Name)
	if existingUser != nil {
		msg := fmt.Sprintf("failed to create user %s: user with that name already exists", user.Name)
		logrus.Error(msg)
		response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: msg}}
	} else {
		err = h.repo.HashPassword(&user, user.Password)
		if err != nil {
			logrus.Error("failed to hash password: " + err.Error())
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to hash password."}}
		} else {
			logrus.Info("Creating new user...")
			uId, err := h.repo.Create(r.Context(), &user)
			if err != nil {
				logrus.Error("failed to create new order: " + err.Error())
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create new order."}}
			} else {
				logrus.Info(fmt.Sprintf("Created new user '%s' with id = %d", user.Name, uId))
				//response = utils.JsonResponse{Id: strconv.Itoa(int(uId))}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else {
		utils.WriteJSONResponse(w, http.StatusNoContent, response)
	}
}

// Login attempts to authorize a user based on the supplied credentials (in the http request), and returns a message in JSON on error, or a JSON Web Token on success.
func (h *User) Login(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodPost, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Login API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}
	logrus.Info("Starting login process...")

	response := utils.JsonResponse{}
	user := models.User{}
	err := utils.DecodeJSONBody(w, r, &user, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if err != nil {
		var request *utils.MalformedRequestError
		if errors.As(err, &request) {
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: request.Status, Message: request.Message}}
		} else {
			msg := "Failed to decode JSON body: " + err.Error()
			logrus.Error(msg)
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: msg}}
		}
	}

	storedUser, err := h.repo.GetByUsername(user.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Error(fmt.Sprintf("failed to find user with username: %s: %s", user.Name, err.Error()))
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "Invalid user credentials."}}
		} else {
			logrus.Error(fmt.Sprintf("failed to retrieve list of users: " + err.Error()))
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Internal server error. Please contact your network administrator."}}
		}
	} else {
		err = h.repo.CheckPassword(storedUser, user.Password)
		if err != nil {
			logrus.Error(fmt.Sprintf("failed to authenticate user %s: password check failed: %s", user.Name, err.Error()))
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "Invalid user credentials."}}
		} else {
			jwt := models.JwtWrapper{
				SecretKey:       SECRET_KEY,
				Issuer:          ISSUER,
				ExpirationHours: 24,
			}
			signedToken, err := h.jwtRepo.GenerateToken(&jwt)
			if err != nil {
				logrus.Error(fmt.Sprintf("failed to sign token: %s", err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to sign token."}}
			} else {
				response = utils.JsonResponse{Token: signedToken}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else {
		logrus.Info(fmt.Sprintf("Completed login for user '%s'", user.Name))
		utils.WriteJSONResponse(w, http.StatusOK, response)
	}
}

// IsAuthorized checks the authorization header of the given HTTP request to see if a valid JSON Web Token is included.
// Returns a status message in JSON on failure.
// On success, moves on to the next http handler in the calling chain.
func (h *User) IsAuthorized(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// don't look for authorization from OPTIONS requests
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		logrus.Info("Starting JWT authorization...")
		auth := r.Header.Get("Authorization")
		if auth == "" {
			logrus.Error("No authorization header provided")
			http.Error(w, "Authorization failed.", http.StatusForbidden)
			return
		}
		token := strings.Split(auth, "Bearer ")
		if len(token) == 2 {
			auth = strings.TrimSpace(token[1])
		} else {
			logrus.Error("Incorrect format of authorization token")
			http.Error(w, "Authorization failed.", http.StatusBadRequest)
			return
		}

		authReal := models.JwtWrapper{
			SecretKey: SECRET_KEY,
			Issuer:    ISSUER,
		}

		_, err := h.jwtRepo.ValidateToken(&authReal, auth)
		if err != nil {
			logrus.Error("Authorization failed: SecretKey and/or Issuer wrong")
			http.Error(w, "Authorization failed.", http.StatusUnauthorized)
			return
		}
		logrus.Info("Authorization successful.")
		next.ServeHTTP(w, r)
	})
}
