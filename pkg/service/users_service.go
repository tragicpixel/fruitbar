package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tragicpixel/fruitbar/pkg/driver"
	pgdriver "github.com/tragicpixel/fruitbar/pkg/driver/postgres"
	"github.com/tragicpixel/fruitbar/pkg/handler"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/models/roles"
	"github.com/tragicpixel/fruitbar/pkg/utils/cors"
	"github.com/tragicpixel/fruitbar/pkg/utils/json"
	"github.com/tragicpixel/fruitbar/pkg/utils/log"
)

// UsersService holds all the pieces necessary to run the authentication service for the fruitbar application.
type UsersService struct {
	Router  *mux.Router
	Handler *handler.User
	DB      *driver.DB
	Port    int
}

type UsersServiceConfig struct {
	DatabaseConnection *pgdriver.PostgresConnectionConfig
	Port               int
}

const (
	usersAPIBaseRoute           = "/users"
	usersCreateAPIRoute         = usersAPIBaseRoute + "/register" // for now, remove /register later
	usersReadAPIRoute           = usersAPIBaseRoute
	usersUpdateAPIRoute         = usersAPIBaseRoute
	usersDeleteAPIRoute         = usersAPIBaseRoute
	usersLoginAPIRoute          = usersAPIBaseRoute + "/login"
	usersPasswordFormatAPIRoute = usersAPIBaseRoute + "/password-format"
	usersHealthAPIRoute         = usersAPIBaseRoute + "/health"
)

func (s *UsersService) getCreateAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Create User",
		AllowedMethods: []string{http.MethodPost, http.MethodOptions},
	}
}
func (s *UsersService) getReadAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Read User",
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	}
}
func (s *UsersService) getUpdateAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Update User",
		AllowedMethods: []string{http.MethodPut, http.MethodOptions},
	}
}
func (s *UsersService) getDeleteAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Delete User",
		AllowedMethods: []string{http.MethodDelete, http.MethodOptions},
	}
}
func (s *UsersService) getLoginAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Login User",
		AllowedMethods: []string{http.MethodPost, http.MethodOptions},
	}
}
func (s *UsersService) getPasswordFormatAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Password Format",
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	}
}
func (s *UsersService) getHealthCheckAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Health Check",
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	}
}

func (s *UsersService) getCreateAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getCreateAPIOptions(), s.Handler.IsAuthorized(s.Handler.HasRole(s.Handler.CreateUser, roles.Admin)))
}
func (s *UsersService) getReadAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getReadAPIOptions(), s.Handler.IsAuthorized(s.Handler.GetUsers))
}
func (s *UsersService) getUpdateAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getUpdateAPIOptions(), s.Handler.IsAuthorized(s.Handler.HasRole(s.Handler.UpdateUser, roles.Admin)))
}
func (s *UsersService) getDeleteAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getDeleteAPIOptions(), s.Handler.IsAuthorized(s.Handler.HasRole(s.Handler.DeleteUser, roles.Admin)))
}
func (s *UsersService) getLoginAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getLoginAPIOptions(), s.Handler.Login)
}
func (s *UsersService) getPasswordFormatAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getPasswordFormatAPIOptions(), s.Handler.GetPasswordFormatMessage)
}
func (s *UsersService) getHealthCheckAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getHealthCheckAPIOptions(), s.CheckHealth)
}

// NewUsersService creates a new instance of a users service.
// Returns nil on error.
func NewUsersService(config *UsersServiceConfig) (*UsersService, error) {
	s := UsersService{}

	// sqldb is service name of postgres container in docker-compose
	db, err := pgdriver.OpenConnection(config.DatabaseConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the user service database: %s", err.Error())
	}

	s.DB = db
	err = SetupUsersServiceDB(s.DB, true)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the user service database: %s", err.Error())
	}

	s.Handler = handler.NewUserHandler(db)
	s.Router = s.NewUsersServiceRouter(db)
	s.Port = config.Port

	return &s, nil
}

// NewUsersServiceRouter creates and returns a new http router for the users service.
func (s *UsersService) NewUsersServiceRouter(db *driver.DB) *mux.Router {
	r := mux.NewRouter()

	// swagger:operation POST /users/login users authUser
	//
	// Log a user in and return a JWT.
	//
	// ---
	// parameters:
	// - name: user
	//   in: body
	//   description: Credentials of the user to verify.
	//   required: true
	//   "$ref": "#/definitions/user"
	// responses:
	//   '200':
	//     description: Successfully logged in.
	//     "$ref": "#/responses/jsonResponse"
	//   '400':
	//     description: Invalid request.
	//     "$ref": "#/responses/jsonResponse"
	//   '401':
	//     description: Not authorized.
	//   '405':
	//     description: HTTP method not allowed.
	//   '413':
	//     description: Request body too large.
	//     "$ref": "#/responses/jsonResponse"
	//   '500':
	//     description: Internal server error.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(usersLoginAPIRoute, s.getLoginAPIHandler()).Methods(s.getLoginAPIOptions().AllowedMethods...)
	// swagger:operation POST /users/register users createUser
	//
	// Create a new user.
	//
	// ---
	// parameters:
	// - name: user
	//   in: body
	//   description: New user to create. Id, CreatedAt, DeletedAt, UpdatedAt fields will be ignored.
	//   required: true
	//   "$ref": "#/definitions/user"
	// responses:
	//   '200':
	//     description: Successfully created a user. Id of the newly created user is not returned.
	//   '400':
	//     description: Invalid request.
	//     "$ref": "#/responses/jsonResponse"
	//   '405':
	//     description: HTTP method not allowed.
	//   '413':
	//     description: Request body too large.
	//     "$ref": "#/responses/jsonResponse"
	//   '500':
	//     description: Internal server error.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(usersCreateAPIRoute, s.getCreateAPIHandler()).Methods(s.getCreateAPIOptions().AllowedMethods...)
	// swagger:operation GET /users users getUser
	//
	// Get a single user by ID, or a paginated array of all users.
	//
	// ---
	// parameters:
	// - name: id
	//   in: query
	//   description: id of user to retrieve.
	//   required: false
	//   schema:
	//     type: int
	// security:
	// - bearer: []
	// responses:
	//   '200':
	//     description: Successfully retrieved a user.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(usersReadAPIRoute, s.getReadAPIHandler()).Methods(s.getReadAPIOptions().AllowedMethods...)
	// swagger:operation PUT /users users updateUser
	//
	// Update an existing uuser.
	//
	// ---
	// parameters:
	// - name: user
	//   in: body
	//   description: user fields to update. CreatedAt, DeletedAt, UpdatedAt fields will be ignored.
	//   required: true
	//   schema:
	//     $ref: "#/definitions/user"
	// security:
	// - bearer: []
	// responses:
	//   '200':
	//     description: Successfully updated an existing user.
	//     "$ref": "#/responses/jsonResponse"
	//   '400':
	//     description: Invalid request.
	//     "$ref": "#/responses/jsonResponse"
	//   '401':
	//     description: Not authorized.
	//   '403':
	//     description: No authorization header provided.
	//   '405':
	//     description: HTTP method not allowed.
	//   '413':
	//     description: Request body too large.
	//     "$ref": "#/responses/jsonResponse"
	//   '500':
	//     description: Internal server error.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(usersUpdateAPIRoute, s.getUpdateAPIHandler()).Methods(s.getUpdateAPIOptions().AllowedMethods...)
	// swagger:operation DELETE /users users deleteUser
	//
	// Delete an existing user.
	//
	// ---
	// parameters:
	// - name: id
	//   in: query
	//   description: id of user to delete.
	//   required: true
	//   schema:
	//     type: int
	// security:
	// - bearer: []
	// responses:
	//   '204':
	//     description: Successfully deleted an existing user.
	//   '400':
	//     description: Invalid request.
	//     "$ref": "#/responses/jsonResponse"
	//   '401':
	//     description: Not authorized.
	//   '403':
	//     description: No authorization header provided.
	//   '405':
	//     description: HTTP method not allowed.
	//   '413':
	//     description: Request body too large.
	//     "$ref": "#/responses/jsonResponse"
	//   '500':
	//     description: Internal server error.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(usersDeleteAPIRoute, s.getDeleteAPIHandler()).Methods(s.getDeleteAPIOptions().AllowedMethods...)
	// swagger:operation GET /users/password-format users getPasswordFormat
	//
	// Returns a string containing information about the expected format of a password.
	//
	// ---
	// responses:
	//   '200':
	//     description: The password format message was returned successfully.
	r.HandleFunc(usersPasswordFormatAPIRoute, s.getPasswordFormatAPIHandler()).Methods(s.getPasswordFormatAPIOptions().AllowedMethods...)
	// swagger:operation GET /users/health users checkHealth
	//
	// Checks the health of the service and sends a response indicating if the health check passed.
	//
	// ---
	// responses:
	//   '200':
	//     description: The health check was completed.
	//     "$ref": "#/responses/healthCheckResponse"
	r.HandleFunc(usersHealthAPIRoute, s.getHealthCheckAPIHandler()).Methods(s.getHealthCheckAPIOptions().AllowedMethods...)

	return r
}

// SetupUsersServiceDB checks that the database schema is ready for the authentication service.
// If init is true, will create the tables if they do not already exist.
func SetupUsersServiceDB(db *driver.DB, init bool) error {
	log.Info("Setting up the users service database...")
	err := pgdriver.SetupTables(db, &models.User{}, init)
	if err != nil {
		log.Error("failed to set up the User model table" + err.Error())
		return errors.New("failed to set up the User model table: " + err.Error())
	}
	log.Info("Successfully set up the database for the users service")
	return nil
}

// CheckHealth checks the health of the authentication service and writes a response in JSON to the user.
func (s *UsersService) CheckHealth(w http.ResponseWriter, r *http.Request) {
	var err error
	log.Info("Checking users service health...")
	db, err := s.DB.Postgres.DB()
	if err != nil {
		log.Error("health check failed: Error getting SQLDB from gorm DB: " + err.Error())
		json.WriteResponse(w, http.StatusOK, map[string]bool{"ok": false})
		return
	}
	if err = db.Ping(); err != nil {
		log.Error("health check failed: error pinging the database: " + err.Error())
		json.WriteResponse(w, http.StatusOK, map[string]bool{"ok": false})
		return
	}
	log.Info("health check passed")
	json.WriteResponse(w, http.StatusOK, map[string]bool{"ok": true})
}
