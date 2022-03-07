package service

import (
	"github.com/tragicpixel/fruitbar/pkg/driver"
	pgdriver "github.com/tragicpixel/fruitbar/pkg/driver/postgres"
	"github.com/tragicpixel/fruitbar/pkg/handler"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils"

	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ProductsService holds all the pieces necessary to run the product listing service for the fruitbar application.
type ProductsService struct {
	Router      *mux.Router
	Handler     *handler.Product
	UserHandler *handler.User
	DB          *driver.DB
	Port        int
}

type ProductsServiceConfig struct {
	DatabaseConnection *pgdriver.PostgresConnectionConfig
	Port               int
}

func (s *ProductsService) getProductsApiBasePath() string  { return "/products" }
func (s *ProductsService) getCreateApiHandlerPath() string { return s.getProductsApiBasePath() }
func (s *ProductsService) getReadApiHandlerPath() string   { return s.getProductsApiBasePath() }
func (s *ProductsService) getUpdateApiHandlerPath() string { return s.getProductsApiBasePath() }
func (s *ProductsService) getDeleteApiHandlerPath() string { return s.getProductsApiBasePath() }
func (s *ProductsService) getHealthCheckApiHandlerPath() string {
	return s.getProductsApiBasePath() + "/health"
}

func (s *ProductsService) getCreateApiAllowedHttpMethods() []string {
	return []string{http.MethodPost, http.MethodOptions}
}
func (s *ProductsService) getReadApiAllowedHttpMethods() []string {
	return []string{http.MethodGet, http.MethodOptions}
}
func (s *ProductsService) getUpdateApiAllowedHttpMethods() []string {
	return []string{http.MethodPut, http.MethodOptions}
}
func (s *ProductsService) getDeleteApiAllowedHttpMethods() []string {
	return []string{http.MethodDelete, http.MethodOptions}
}
func (s *ProductsService) getHealthCheckApiAllowedHttpMethods() []string {
	return []string{http.MethodGet, http.MethodOptions}
}

// NewProductsService creates a new instance of a product listing service.
// Returns nil on error.
func NewProductsService(config *ProductsServiceConfig) (*ProductsService, error) {
	s := ProductsService{}

	db, err := pgdriver.OpenConnection(config.DatabaseConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the products service database: %s", err.Error())
	}

	s.DB = db
	err = setupProductsServiceDB(s.DB, true)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the products service database: %s", err.Error())
	}

	s.Handler = handler.NewProductHandler(db)
	s.UserHandler = handler.NewUserHandler(db)
	s.Router = s.NewProductsServiceRouter(db)
	s.Port = config.Port

	return &s, nil
}

// NewProductsServiceRouter creates and returns a new http router for the product listing service.
func (s *ProductsService) NewProductsServiceRouter(db *driver.DB) *mux.Router {
	r := mux.NewRouter()

	// swagger:operation POST /products products createProduct
	//
	// Create a new product.
	//
	// ---
	// parameters:
	// - name: product
	//   in: body
	//   description: New product to create. Id, CreatedAt, DeletedAt, UpdatedAt fields will be ignored.
	//   required: true
	//   "$ref": "#/definitions/product"
	// security:
	// - bearer: []
	// responses:
	//   '201':
	//     description: Successfully created a product.
	//     "$ref": "#/responses/jsonResponse"
	//     examples:
	//       application/json: { "ok": 2 }
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
	r.HandleFunc(s.getCreateApiHandlerPath(), s.UserHandler.IsAuthorized(s.Handler.CreateProduct)).Methods(s.getCreateApiAllowedHttpMethods()...)
	// swagger:operation GET /products products getProduct
	//
	// Get a product by ID, or a paginated listing of all products.
	//
	// ---
	// parameters:
	// - name: id
	//   in: query
	//   description: id of product to retrieve.
	//   required: false
	//   schema:
	//     type: int
	// security:
	// - bearer: []
	// responses:
	//   '200':
	//     description: Successfully retrieved a product.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(s.getReadApiHandlerPath(), s.UserHandler.IsAuthorized(s.Handler.GetProducts)).Methods(s.getReadApiAllowedHttpMethods()...)
	// swagger:operation PUT /products products updateProduct
	//
	// Update an existing product.
	//
	// ---
	// parameters:
	// - name: product
	//   in: body
	//   description: Product fields to update. CreatedAt, DeletedAt, UpdatedAt fields will be ignored.
	//   required: true
	//   schema:
	//     $ref: "#/definitions/product"
	// security:
	// - bearer: []
	// responses:
	//   '200':
	//     description: Successfully updated an existing product.
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
	r.HandleFunc(s.getUpdateApiHandlerPath(), s.UserHandler.IsAuthorized(s.Handler.UpdateProduct)).Methods(s.getUpdateApiAllowedHttpMethods()...)
	// swagger:operation DELETE /products products deleteProduct
	//
	// Delete an existing product.
	//
	// ---
	// parameters:
	// - name: id
	//   in: query
	//   description: id of product to delete.
	//   required: true
	//   schema:
	//     type: int
	// security:
	// - bearer: []
	// responses:
	//   '204':
	//     description: Successfully deleted an existing product.
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
	r.HandleFunc(s.getDeleteApiHandlerPath(), s.UserHandler.IsAuthorized(s.Handler.DeleteProduct)).Methods(s.getDeleteApiAllowedHttpMethods()...)
	// swagger:operation GET /products/health products checkHealth
	//
	// Checks the health of the service.
	//
	// ---
	// responses:
	//   '200':
	//     description: The health check was completed.
	//     "$ref": "#/responses/healthCheckResponse"
	r.HandleFunc(s.getHealthCheckApiHandlerPath(), s.CheckHealth).Methods(s.getHealthCheckApiAllowedHttpMethods()...)

	return r
}

// CheckHealth checks the health of the product listing service and writes a response in JSON to the user.
// Always returns HTTP Status OK, even if the health check fails.
func (s *ProductsService) CheckHealth(w http.ResponseWriter, r *http.Request) {
	var err error
	logrus.Info("Checking products service health...")
	theDatabase, err := s.DB.Postgres.DB()
	if err != nil {
		logrus.Error("products service health check failed: Error getting SQLDB from gorm DB: " + err.Error())
		utils.WriteJSONResponse(w, http.StatusOK, map[string]bool{"ok": false})
	} else {
		if err = theDatabase.Ping(); err != nil {
			logrus.Error("products service health check failed: error pinging the database: " + err.Error())
			utils.WriteJSONResponse(w, http.StatusOK, map[string]bool{"ok": false})
		} else {
			logrus.Info("products service health check passed")
			utils.WriteJSONResponse(w, http.StatusOK, map[string]bool{"ok": true})
		}
	}
}

// SetupOrdersServiceDB checks that the database schema is ready for the authentication service.
// If init is true, will create the tables if they do not already exist.
func setupProductsServiceDB(db *driver.DB, init bool) error {
	logrus.Info("Setting up the products service database...")
	err := pgdriver.SetupTables(db, &models.Product{}, init)
	if err != nil {
		msg := "failed to set up the Product model table" + err.Error()
		logrus.Error(msg)
		return errors.New(msg)
	}
	logrus.Info("Successfully set up the database for the products service")
	return nil
}
