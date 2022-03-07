package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/tragicpixel/fruitbar/pkg/driver"
	pgdriver "github.com/tragicpixel/fruitbar/pkg/driver/postgres"
	"github.com/tragicpixel/fruitbar/pkg/handler"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils"
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
	usersServiceRegisterApiHandlerPath    = "/users/register"
	usersServiceLoginApiHandlerPath       = "/users/login"
	usersServiceHealthCheckApiHandlerPath = "/users/health"
)

func getUsersRegisterApiAllowedHttpMethods() []string {
	return []string{http.MethodPost, http.MethodOptions}
}
func getUsersLoginApiAllowedHttpMethods() []string {
	return []string{http.MethodPost, http.MethodOptions}
}
func getUsersHealthCheckApiAllowedHttpMethods() []string {
	return []string{http.MethodGet, http.MethodOptions}
}

// NewUsersService creates a new instance of an authentication service.
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

// NewUsersServiceRouter creates and returns a new http router for the authentication service.
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
	r.HandleFunc(usersServiceLoginApiHandlerPath, s.Handler.Login).Methods(getUsersLoginApiAllowedHttpMethods()...)
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
	r.HandleFunc(usersServiceRegisterApiHandlerPath, s.Handler.RegisterNewUser).Methods(getUsersRegisterApiAllowedHttpMethods()...)
	// swagger:operation GET /users/health users checkHealth
	//
	// Checks the health of the service.
	//
	// ---
	// responses:
	//   '200':
	//     description: The health check was completed.
	//     "$ref": "#/responses/healthCheckResponse"
	r.HandleFunc(usersServiceHealthCheckApiHandlerPath, s.CheckHealth).Methods(getUsersHealthCheckApiAllowedHttpMethods()...)

	return r
}

// SetupUsersServiceDB checks that the database schema is ready for the authentication service.
// If init is true, will create the tables if they do not already exist.
func SetupUsersServiceDB(db *driver.DB, init bool) error {
	logrus.Info("Setting up the users service database...")
	err := pgdriver.SetupTables(db, &models.User{}, init)
	if err != nil {
		logrus.Error("failed to set up the User model table" + err.Error())
		return errors.New("failed to set up the User model table: " + err.Error())
	}
	logrus.Info("Successfully set up the database for the users service")
	return nil
}

// CheckHealth checks the health of the authentication service and writes a response in JSON to the user.
func (s *UsersService) CheckHealth(w http.ResponseWriter, r *http.Request) {
	var err error
	logrus.Info("Checking users service health...")
	theDatabase, err := s.DB.Postgres.DB()
	if err != nil {
		logrus.Error("health check failed: Error getting SQLDB from gorm DB: " + err.Error())
		utils.WriteJSONResponse(w, http.StatusOK, map[string]bool{"ok": false})
	} else {
		if err = theDatabase.Ping(); err != nil {
			logrus.Error("health check failed: error pinging the database: " + err.Error())
			utils.WriteJSONResponse(w, http.StatusOK, map[string]bool{"ok": false})
		} else {
			logrus.Info("health check passed")
			utils.WriteJSONResponse(w, http.StatusOK, map[string]bool{"ok": true})
		}
	}
}
