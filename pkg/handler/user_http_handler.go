package handler

import (
	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
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

// TODO: How to properly store these???
const (
	UI_URL                    = "http://localhost:3000"
	forbiddenErrMsg           = "Forbidden: Not enough privileges to "
	forbiddenCreateUserErrMsg = forbiddenErrMsg + "create Users with the 'employee' or 'admin' roles."
	forbiddenReadUserErrMsg   = forbiddenErrMsg + "read this User."
	forbiddenUpdateUserErrMsg = forbiddenErrMsg + "update this User."
	forbiddenDeleteUserErrMsg = forbiddenErrMsg + "delete this User."
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

// CreateUser creates a new user with the specified username and password, and returns a JSON message if there is an error, otherwise no content.
func (h *User) CreateUser(w http.ResponseWriter, r *http.Request) {
	var response = utils.JsonResponse{}
	var user models.User
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &user, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	_, err := models.ValidateRole(user.Role)
	if err != nil {
		msg := "Role is invalid: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, msg)
		return
	}

	requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
		return
	}
	if user.Role != models.GetCustomerRoleId() && requestor.UserRole != models.GetAdminRoleId() {
		http.Error(w, forbiddenCreateUserErrMsg, http.StatusForbidden)
		return
	}

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
	response = utils.JsonResponse{Data: id} // TODO: Figure out the new format for returning ids/data and change this
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetUsers retrieves either a single users in the database if the "id" query parameter is supplied, or a list of users in the database if it is not.
// The list of products can also be paginated, using the "limit" and "after_id" query parameters.
// Returns a response in JSON containing either an array of users encoded in JSON or an error message.
func (h *User) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has(idParam) {
		h.getSingleUser(w, r)
	} else {
		h.getUsersPage(w, r)
	}
}

// getSingleUser retrieves a single user from the user repository based on the supplied id via http query parameter.
// Sends a response in json to the supplied http ResponseWriter.
func (h *User) getSingleUser(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	logrus.Info(fmt.Sprintf("Reading user (id: %d)...", id))
	var user *models.User
	requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
		return
	}
	_, err = models.ValidateRole(requestor.UserRole)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusUnauthorized, "Authorization Failed", logMsg)
		return
	}

	// Customers can only read their own user account
	if requestor.UserRole == models.GetCustomerRoleId() && requestor.UserID != id {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenReadUserErrMsg)
		return
	}

	user, err = h.repo.GetByID(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading user (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Read user (id: %d)", id))
	user.Password = "" // Remove password hash for security reasons

	// Employees can read any customer user account and their own user account
	if requestor.UserRole == models.GetEmployeeRoleId() && (user.Role != models.GetCustomerRoleId() && requestor.UserID != id) {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenReadUserErrMsg)
		return
	}

	response := utils.JsonResponse{Data: []*models.User{user}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// getUsersPage retrieves a single user from the user repository based on the supplied seek options via http query parameter.
// Sends a response in json to the supplied http ResponseWriter.
func (h *User) getUsersPage(w http.ResponseWriter, r *http.Request) {
	var seek *utils.PageSeekOptions
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

	requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
		return
	}
	var pruned []*models.User
	switch requestor.UserRole {
	case models.GetAdminRoleId():
		// Admin can read all users
		pruned = users
	case models.GetEmployeeRoleId():
		// Employees can only read: other customer users, and their user
		for _, user := range users {
			if user.ID == requestor.UserID || user.Role == models.GetCustomerRoleId() {
				pruned = append(pruned, user)
			}
		}
	case models.GetCustomerRoleId():
		// Customers can only read: their user
		for _, user := range users {
			if user.ID == requestor.UserID {
				pruned = append(pruned, user)
			}
		}
	}
	users = pruned

	count, err := h.repo.Count(seek)
	if err != nil {
		logMsg := fmt.Sprintf("Error counting users: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	startID, endID := uint(0), uint(0)
	if len(users) > 0 {
		startID = users[0].ID
		endID = users[len(users)-1].ID
	}
	rangeStr := fmt.Sprintf("users=%d-%d/%d", startID, endID, count)
	w.Header().Set("Content-Range", rangeStr)
	logrus.Info(fmt.Sprintf("Read %d users", len(users)))
	response := utils.JsonResponse{Data: users}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// UpdateProduct updates an existing product in the repo based on the supplied JSON request, and returns a status message in JSON to the user.
// If price is not set or set to zero, it will be ignored.
func (h *User) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var response = utils.JsonResponse{}
	var user models.User
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &user, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
		return
	}
	// Customers can only update their own user account
	if requestor.UserRole == models.GetCustomerRoleId() && user.ID != requestor.UserID {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenUpdateUserErrMsg)
		return
	}
	// Employees can only update customer accounts and their own user account
	if requestor.UserRole == models.GetEmployeeRoleId() && (user.ID != requestor.UserID && user.Role != models.GetCustomerRoleId()) {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenUpdateUserErrMsg)
		return
	}

	if r.URL.Query().Has(fieldsParam) {
		h.partiallyUpdateUser(w, r, user)
	} else {
		h.fullyUpdateUser(w, r, user)
	}
}

// partiallyUpdateUser updates only the specified fields (via http query parameter) of the supplied user.
// Sends a response in json to the supplied http ResponseWriter.
func (h *User) partiallyUpdateUser(w http.ResponseWriter, r *http.Request, user models.User) {
	fieldsStr := r.URL.Query().Get(fieldsParam)
	fields := strings.Split(fieldsStr, ",")

	_, err := models.ValidatePartialUserUpdate(&user, fields)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "User "+validationFailedMsg+err.Error())
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

// fullyUpdateUser updates all of the fields of the supplied user.
// Sends a response in json to the supplied http ResponseWriter.
func (h *User) fullyUpdateUser(w http.ResponseWriter, r *http.Request, user models.User) {
	_, err := models.ValidateUser(&user)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "User "+validationFailedMsg+err.Error())
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

// DeleteProduct deletes an existing user from the repo based on the supplied http request, and returns a status message in JSON to the user.
func (h *User) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
		return
	}
	// Prevent users from deleting themselves.
	if requestor.UserID == id {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenDeleteUserErrMsg)
		return
	}
	// Customers can only update their own user account
	if requestor.UserRole == models.GetCustomerRoleId() && id != requestor.UserID {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenDeleteUserErrMsg)
		return
	}
	// Employees can only update customer accounts and their own user account
	if requestor.UserRole == models.GetEmployeeRoleId() && id != requestor.UserID {
		logrus.Info("Reading User for proposed delete...")
		user, err := h.repo.GetByID(id)
		if err != nil {
			logMsg := fmt.Sprintf("Error reading user (id: %d): %s", id, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
		if user.Role != models.GetCustomerRoleId() {
			utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenDeleteUserErrMsg)
			return
		}
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
	utils.WriteJSONResponse(w, http.StatusOK, utils.JsonResponse{Data: models.GetPasswordFormatMessage()})
}

// Login attempts to authorize a user based on the supplied credentials (in the http request), and returns a message in JSON on error, or a JSON Web Token on success.
func (h *User) Login(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	err := utils.DecodeJSONBody(w, r, &user, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if err != nil {
		var request *utils.MalformedRequestError
		if errors.As(err, &request) {
			utils.WriteJSONErrorResponse(w, request.Status, request.Message)
			return
		}
		logMsg := fmt.Sprintf("Failed to decode JSON body: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
	}

	storedUser, err := h.repo.GetByUsername(user.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logMsg := fmt.Sprintf("failed to find user with username: %s: %s", user.Name, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Invalid user credentials.", logMsg)
			return
		}
		logrus.Error(fmt.Sprintf("failed to retrieve list of users: " + err.Error()))
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
	jwt.ExpirationHours = 24 // can I put this in GetRealAuthToken? test

	signedToken, err := h.jwtRepo.GenerateToken(&jwt, storedUser)
	if err != nil {
		logrus.Error(fmt.Sprintf("Failed to generate token: %s", err.Error()))
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Authentication successful for user '%s'", user.Name))
	response := utils.JsonResponse{Token: signedToken}
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

		logrus.Info("Starting JWT authorization...")
		auth := r.Header.Get("Authorization")
		if auth == "" {
			logrus.Error("No authorization header provided")
			http.Error(w, "Authorization failed.", http.StatusForbidden)
			return
		}

		authToken, err := jwtutils.GetTokenFromAuthHeader(auth)
		if err != nil {
			logrus.Error(err.Error())
			http.Error(w, "Authorization failed.", http.StatusBadRequest)
			return
		}

		authReal := jwtutils.GetSecretAuthToken()

		_, err = h.jwtRepo.ValidateToken(&authReal, authToken)
		if err != nil {
			logrus.Error("Authorization failed: SecretKey and/or Issuer wrong")
			http.Error(w, "Authorization failed.", http.StatusUnauthorized)
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
			logMsg := "Authorization failed: " + err.Error()
			utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
			return
		}

		hasRole, err := models.HasRole(requestor.UserRole, role)
		if err != nil {
			logrus.Error(fmt.Sprintf("Unexpected error checking user's role: %s", err.Error()))
			http.Error(w, internalServerErrMsg, http.StatusInternalServerError)
			return
		}

		if !hasRole {
			logrus.Error(fmt.Sprintf("User's role does meet the access level requirements: expecting '%s' got '%s'", role, requestor.UserRole))
			http.Error(w, "Authorization failed.", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
