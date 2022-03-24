package handler

import (
	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/models/roles"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	jwtrepo "github.com/tragicpixel/fruitbar/pkg/repository/jwt"
	userrepo "github.com/tragicpixel/fruitbar/pkg/repository/user"
	"github.com/tragicpixel/fruitbar/pkg/utils"
	jwtutils "github.com/tragicpixel/fruitbar/pkg/utils/jwt"
	stringutils "github.com/tragicpixel/fruitbar/pkg/utils/string"
	"gorm.io/gorm"

	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// User represents a handler for performing operations on users via HTTP.
type User struct {
	repo    repository.User
	jwtRepo repository.Jwt
}

// NewUserHandler creates and initializes a new handler for performing operations on users via HTTP.
func NewUserHandler(db *driver.DB) *User {
	return &User{
		repo:    userrepo.NewPostgresUserRepo(db.Postgres), // this is where it is decided which implementation(/database type) of the User Repo we will use
		jwtRepo: jwtrepo.NewJWTRepository(),
	}
}

// CreateUser creates a new user based on the supplied HTTP request and sends a response in JSON containing the newly created user to the supplied http response writer.
// If there is a permission error, an HTTP error will be sent.
func (h *User) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	response := *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &user, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}
	_, err := user.IsValid()
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "User "+validationFailedErrMsgPrefix+err.Error())
		return
	}

	if !h.clientHasCreateUserPermsForUser(w, r, user) {
		return
	}

	logrus.Info(fmt.Sprintf("Checking if user %s exists...", user.Name))
	existingUser, err := h.repo.GetByUsername(user.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logMsg := fmt.Sprintf("Failed to check if user %s exists: %s", user.Name, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	if existingUser != nil {
		msg := fmt.Sprintf("Failed to create user %s: a user with that name already exists", user.Name)
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, msg)
		return
	}

	err = h.repo.HashPassword(&user, user.Password)
	if err != nil {
		logMsg := fmt.Sprintf("failed to hash password: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}

	logrus.Info("Creating new user...")
	id, err := h.repo.Create(&user)
	if err != nil {
		logMsg := fmt.Sprintf("Failed to create new user %s: %s", user.Name, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Created new user '%s' (id: %d)", user.Name, id))
	response = utils.JsonResponse{Data: []*models.User{&user}}
	utils.WriteJSONResponse(w, http.StatusCreated, response)
}

// GetUsers sends a response to the supplied http response writer containing the requested user(s), based on the supplied http request.
func (h *User) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has(idParam) {
		h.getSingleUser(w, r)
	} else {
		h.getUsersPage(w, r)
	}
}

// UpdateUser updates an existing user based on the supplied http request and sends a response in JSON containing the updated user to the supplied http response writer.
func (h *User) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	response := *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &user, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	if !h.clientHasUpdatePermsForUser(w, r, user) {
		return
	}

	if r.URL.Query().Has(fieldsParam) {
		h.partiallyUpdateUser(w, r, user)
	} else {
		h.fullyUpdateUser(w, r, user)
	}
}

// DeleteProduct deletes an existing user from the repo based on the supplied http request, and returns a status message in JSON to the user.
func (h *User) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if !h.clientHasDeletePermsForID(w, r, id) {
		return
	}

	logrus.Info(fmt.Sprintf("Deleting User (id: %d)...", id))
	err = h.repo.Delete(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error deleting User (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Successfully deleted User with id = %d.", id))
	response := utils.JsonResponse{}
	utils.WriteJSONResponse(w, http.StatusOK, response) // could also just w.WriteHeader(http.StatusNoContent) and not send a response at all?
}

// GetPasswordFormatMessage always sends a response containing a message explaining the constraints applied when setting a new password.
// This is the single source of truth for information about the expected format of a new password.
func (h *User) GetPasswordFormatMessage(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSONResponse(w, http.StatusOK, utils.JsonResponse{Data: models.PasswordFormatReqMsg()})
}

// Login attempts to authorize a user based on the supplied credentials (in the http request), and returns a message in JSON on error, or a JSON Web Token on success.
func (h *User) Login(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	response := *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &user, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	logrus.Info(fmt.Sprintf("Selecting user '%s' for login...", user.Name))
	storedUser, err := h.repo.GetByUsername(user.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logMsg := fmt.Sprintf("failed to find user with username: %s: %s", user.Name, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Invalid user credentials.", logMsg)
			return
		}
		logrus.Error(fmt.Sprintf("failed to select user '%s' for login: %s", user.Name, err.Error()))
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg)
		return
	}

	err = h.repo.CheckPassword(storedUser, user.Password)
	if err != nil {
		logrus.Error(fmt.Sprintf("Failed to authenticate user %s: password check failed: %s", user.Name, err.Error()))
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Invalid user credentials.")
		return
	}

	jwt := jwtutils.GetSecretAuthToken()
	jwt.ExpirationHours = 24 // TODO: can I put this in GetSecretAuthToken? test, might not be able to if every token's expiration doesn't get set

	signedToken, err := h.jwtRepo.GenerateToken(&jwt, storedUser)
	if err != nil {
		logrus.Error(fmt.Sprintf("Failed to generate token: %s", err.Error()))
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Authentication successful for user '%s'", user.Name))
	response = utils.JsonResponse{Token: signedToken}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// IsAuthorized checks the authorization header of the given HTTP request to see if a valid JSON Web Token is included.
// Returns a status message in JSON on failure.
// On success, moves on to the next http handler in the calling chain.
func (h *User) IsAuthorized(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions { // don't look for authorization from OPTIONS requests
			next.ServeHTTP(w, r)
			return
		}

		logrus.Info("Starting client authorization...")
		auth := r.Header.Get("Authorization")
		if auth == "" {
			logrus.Error(unauthorizedErrMsgPrefix + "No authorization header provided")
			http.Error(w, unauthorizedErrMsg, http.StatusForbidden)
			return
		}

		authToken, err := jwtutils.GetTokenFromAuthHeader(auth)
		if err != nil {
			logrus.Error(err.Error())
			http.Error(w, unauthorizedErrMsg, http.StatusBadRequest)
			return
		}

		authReal := jwtutils.GetSecretAuthToken()

		_, err = h.jwtRepo.ValidateToken(&authReal, authToken)
		if err != nil {
			logrus.Error(unauthorizedErrMsgPrefix + "SecretKey and/or Issuer wrong")
			http.Error(w, unauthorizedErrMsg, http.StatusUnauthorized)
			return
		}
		logrus.Info("Authorization successful.")
		next.ServeHTTP(w, r)
	})
}

// HasRole checks if the client's JWT contains the specified role. A user's role is specified in the claims in their JWT when they log in.
func (h *User) HasRole(next http.HandlerFunc, role string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't check for roles on OPTIONS requests
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
		if err != nil {
			logMsg := unauthorizedErrMsgPrefix + err.Error()
			utils.WriteJSONErrorResponse(w, http.StatusBadRequest, unauthorizedErrMsg, logMsg)
			return
		}

		hasRole, err := roles.HasRole(requestor.UserRole, role)
		if err != nil {
			logrus.Error(fmt.Sprintf("Unexpected error checking client's role: %s", err.Error()))
			http.Error(w, internalServerErrMsg, http.StatusInternalServerError)
			return
		}

		if !hasRole {
			logrus.Error(fmt.Sprintf(unauthorizedErrMsgPrefix+"Client's role does not meet the access level requirements: expecting '%s' got '%s'", role, requestor.UserRole))
			http.Error(w, unauthorizedErrMsg, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// getSingleUser retrieves a single user from the user repository based on the supplied id via http query parameter.
// Sends a response in json to the supplied http ResponseWriter.
func (h *User) getSingleUser(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if !h.clientHasReadUserPermsForID(w, r, id) {
		return
	}

	var user *models.User
	logrus.Info(fmt.Sprintf("Selecting user (id: %d)...", id))
	user, err = h.repo.GetByID(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error selecting user (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Read user (id: %d)", id))
	user.Password = "" // Remove password hash for security reasons

	if !h.clientHasReadPermsForUser(w, r, user) {
		return
	}

	response := utils.JsonResponse{Data: []*models.User{user}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// getUsersPage retrieves a single user from the user repository based on the supplied seek options via http query parameter.
// Sends a response in json to the supplied http ResponseWriter.
func (h *User) getUsersPage(w http.ResponseWriter, r *http.Request) {
	var seek *repository.PageSeekOptions
	seek, err := utils.GetPageSeekOptions(r, readPageMaxLimit)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	logrus.Info(fmt.Sprintf("Reading %d users (max %d)...", seek.RecordLimit, readPageMaxLimit))
	var users []*models.User
	users, err = h.repo.Fetch(seek)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading users: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	for _, user := range users {
		user.Password = "" // Remove password hash for security reasons
	}

	users = h.getUsersReadableByClient(w, r, users)
	if users == nil {
		return
	}

	rangeStr := h.getUsersRangeStr(w, seek, users)
	w.Header().Set("Content-Range", rangeStr)

	logrus.Info(fmt.Sprintf("Read %d users", len(users)))
	response := utils.JsonResponse{Data: users}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// partiallyUpdateUser updates only the specified fields (via http query parameter) of the supplied user
// and sends a response to the supplied http response writer containing the updated user in JSON.
func (h *User) partiallyUpdateUser(w http.ResponseWriter, r *http.Request, user models.User) {
	fieldsStr := r.URL.Query().Get(fieldsParam)
	fields := strings.Split(fieldsStr, ",")

	_, err := user.ValidatePartialUserUpdate(fields)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "User "+validationFailedErrMsgPrefix+err.Error())
		return
	}

	if stringutils.IsStringInSlice("password", fields) {
		logrus.Info(fmt.Sprintf("Password changed for user %s: Hashing password...", user.Name))
		err = h.repo.HashPassword(&user, user.Password)
		if err != nil {
			logMsg := fmt.Sprintf("Failed to hash password: %s", err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
	}

	logrus.Info(fmt.Sprintf("Updating User (id: %d) fields (%s) to %+v", user.ID, fieldsStr, user))
	updated, err := h.repo.Update(&user, fields)
	if err != nil {
		logMsg := fmt.Sprintf("Error partially updating User (id: %d)  fields (%s) : %s", user.ID, fieldsStr, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Partially updated User (id: %d) fields (%s): %+v", user.ID, fieldsStr, updated))
	response := utils.JsonResponse{Data: []*models.User{&user}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// fullyUpdateUser updates all of the fields of the supplied user
// and sends a response to the supplied http response writer containing the updated user in JSON.
func (h *User) fullyUpdateUser(w http.ResponseWriter, r *http.Request, user models.User) {
	_, err := user.IsValid()
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "User "+validationFailedErrMsgPrefix+err.Error())
		return
	}

	err = h.repo.HashPassword(&user, user.Password)
	if err != nil {
		logMsg := fmt.Sprintf("Failed to hash password: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}

	logrus.Info(fmt.Sprintf("Updating User (id: %d) to %+v", user.ID, user))
	updated, err := h.repo.Update(&user, []string{})
	if err != nil {
		logMsg := fmt.Sprintf("Error fully updating User with id = %d: %+v: %s", user.ID, user, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Fully updated User (id: %d): %+v", user.ID, updated))
	response := utils.JsonResponse{Data: []*models.User{&user}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// getClientAuthInfo returns the authorization information about the client based on the supplied http request.
// Writes a response on the supplied http writer if there is an error.
func (h *User) getClientAuthInfo(w http.ResponseWriter, r *http.Request) *models.JwtClaim {
	client, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := unauthorizedErrMsgPrefix + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, unauthorizedErrMsg, logMsg)
		return nil
	}
	_, err = roles.IsValid(client.UserRole)
	if err != nil {
		logMsg := unauthorizedErrMsgPrefix + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusUnauthorized, unauthorizedErrMsg, logMsg)
		return nil
	}
	return client
}

// clientHasCreatePermsForUser checks whether the client has permissions to create the supplied user, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *User) clientHasCreateUserPermsForUser(w http.ResponseWriter, r *http.Request, user models.User) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Only admins can create employee or admin users
	if user.Role != roles.Customer && client.UserRole != roles.Admin {
		http.Error(w, forbiddenCreateUserErrMsg, http.StatusForbidden)
		return false
	}
	return true
}

// clientHasReadPerms checks whether the client has permissions to read the supplied user or user with the supplied id, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *User) clientHasReadPerms(w http.ResponseWriter, r *http.Request, id uint, user *models.User) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	if user == nil {
		// Customers can only read their own user account
		if client.UserRole == roles.Customer && client.UserID != id {
			utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenReadUserErrMsg)
			return false
		}
	} else {
		// Employees can read any customer user account and their own user account
		if client.UserRole == roles.Employee && (user.Role != roles.Customer && client.UserID != id) {
			utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenReadUserErrMsg)
			return false
		}
	}
	return true
}

// clientHasReadPermsForID checks whether the client has permissions to read the user with the supplied ID, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *User) clientHasReadUserPermsForID(w http.ResponseWriter, r *http.Request, id uint) bool {
	return h.clientHasReadPerms(w, r, id, nil)
}

// clientHasReadPermsForUser checks whether the client has permissions to read the supplied user, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *User) clientHasReadPermsForUser(w http.ResponseWriter, r *http.Request, user *models.User) bool {
	return h.clientHasReadPerms(w, r, 0, user)
}

// clientHasCreatePermsForOrder prunes the supplied users down to only the users the client has permission to read.
func (h *User) getUsersReadableByClient(w http.ResponseWriter, r *http.Request, users []*models.User) []*models.User {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return nil
	}
	var pruned []*models.User
	switch client.UserRole {
	case roles.Admin:
		// Admin can read all users
		pruned = users
	case roles.Employee:
		// Employees can only read: other customer users, and their user
		for _, user := range users {
			if user.ID == client.UserID || user.Role == roles.Customer {
				pruned = append(pruned, user)
			}
		}
	case roles.Customer:
		// Customers can only read: their user
		for _, user := range users {
			if user.ID == client.UserID {
				pruned = append(pruned, user)
			}
		}
	}
	return pruned
}

// clientHasUpdatePermsForUser checks whether the client has permissions to update the supplied user, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *User) clientHasUpdatePermsForUser(w http.ResponseWriter, r *http.Request, user models.User) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Customers can only update their own user account
	if client.UserRole == roles.Customer && user.ID != client.UserID {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenUpdateUserErrMsg)
		return false
	}
	// Employees can only update customer accounts and their own user account
	if client.UserRole == roles.Employee && (user.ID != client.UserID && user.Role != roles.Customer) {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenUpdateUserErrMsg)
		return false
	}
	return true
}

// clientHasDeletePermsForID checks whether the client has permissions to delete a user with the supplied ID, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *User) clientHasDeletePermsForID(w http.ResponseWriter, r *http.Request, id uint) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Prevent users from deleting themselves.
	if client.UserID == id {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenDeleteUserErrMsg)
		return false
	}
	// Customers can only update their own user account
	if client.UserRole == roles.Customer && id != client.UserID {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenDeleteUserErrMsg)
		return false
	}
	// Employees can only update customer accounts and their own user account
	if client.UserRole == roles.Employee && id != client.UserID {
		logrus.Info("Reading User for proposed delete...")
		user, err := h.repo.GetByID(id)
		if err != nil {
			logMsg := fmt.Sprintf("Error reading user (id: %d): %s", id, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return false
		}
		if user.Role != roles.Customer {
			utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenDeleteUserErrMsg)
			return false
		}
	}
	return true
}

// getUsersRangeStr returns a string representation of the range of the supplied products.
func (h *User) getUsersRangeStr(w http.ResponseWriter, seek *repository.PageSeekOptions, users []*models.User) string {
	logrus.Info("Counting users for users page read...")
	count, err := h.repo.Count(seek)
	if err != nil {
		logMsg := fmt.Sprintf("Error counting users: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return ""
	}
	startID, endID := uint(0), uint(0)
	if len(users) > 0 {
		startID = users[0].ID
		endID = users[len(users)-1].ID
	}
	rangeStr := fmt.Sprintf("users=%d-%d/%d", startID, endID, count)
	return rangeStr
}
