package service

import (
	"github.com/tragicpixel/fruitbar/pkg/driver"
	pgdriver "github.com/tragicpixel/fruitbar/pkg/driver/postgres"
	"github.com/tragicpixel/fruitbar/pkg/handler"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/models/roles"
	"github.com/tragicpixel/fruitbar/pkg/utils/cors"
	"github.com/tragicpixel/fruitbar/pkg/utils/json"
	"github.com/tragicpixel/fruitbar/pkg/utils/log"

	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
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

const (
	productsAPIBaseRoute               = "/products"
	productsCreateAPIRoute             = productsAPIBaseRoute
	productsReadAPIRoute               = productsAPIBaseRoute
	productsUpdateAPIRoute             = productsAPIBaseRoute
	productsDeleteAPIRoute             = productsAPIBaseRoute
	productsPageMaxRecordLimitAPIRoute = productsAPIBaseRoute + "/page-max-record-limit"
	productsHealthAPIRoute             = productsAPIBaseRoute + "/health"
)

func (s *ProductsService) getProductsEndpointOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Products Options",
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}
}
func (s *ProductsService) getCreateAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Create Product",
		AllowedMethods: []string{http.MethodPost},
	}
}
func (s *ProductsService) getReadAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Read Product",
		AllowedMethods: []string{http.MethodGet},
	}
}
func (s *ProductsService) getUpdateAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Update Product",
		AllowedMethods: []string{http.MethodPut},
	}
}
func (s *ProductsService) getDeleteAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Delete Product",
		AllowedMethods: []string{http.MethodDelete},
	}
}
func (s *ProductsService) getPageMaxRecordLimitAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Page Max Record Limit",
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	}
}
func (s *ProductsService) getHealthCheckAPIOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Health Check",
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	}
}

func (s *ProductsService) getCreateAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getCreateAPIOptions(), s.UserHandler.IsAuthorized(s.UserHandler.HasRole(s.Handler.CreateProduct, roles.Admin)))
}
func (s *ProductsService) getReadAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getReadAPIOptions(), s.UserHandler.IsAuthorized(s.Handler.GetProducts))
}
func (s *ProductsService) getUpdateAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getUpdateAPIOptions(), s.UserHandler.IsAuthorized(s.UserHandler.HasRole(s.Handler.UpdateProduct, roles.Admin)))
}
func (s *ProductsService) getDeleteAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getDeleteAPIOptions(), s.UserHandler.IsAuthorized(s.UserHandler.HasRole(s.Handler.DeleteProduct, roles.Admin)))
}
func (s *ProductsService) getPageMaxRecordLimitAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getPageMaxRecordLimitAPIOptions(), s.Handler.GetPageMaxRecordLimit)
}
func (s *ProductsService) getHealthCheckAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getHealthCheckAPIOptions(), s.CheckHealth)
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

	r.HandleFunc(productsAPIBaseRoute, cors.SendPreflightHeaders(s.getProductsEndpointOptions(), nil)).Methods(http.MethodOptions)
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
	r.HandleFunc(productsCreateAPIRoute, s.getCreateAPIHandler()).Methods(s.getCreateAPIOptions().AllowedMethods...)
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
	r.HandleFunc(productsReadAPIRoute, s.getReadAPIHandler()).Methods(s.getReadAPIOptions().AllowedMethods...)
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
	r.HandleFunc(productsUpdateAPIRoute, s.getUpdateAPIHandler()).Methods(s.getUpdateAPIOptions().AllowedMethods...)
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
	r.HandleFunc(productsDeleteAPIRoute, s.getDeleteAPIHandler()).Methods(s.getDeleteAPIOptions().AllowedMethods...)
	// swagger:operation GET /products/page-max-record-limit products getPageMaxRecordLimit
	//
	// Returns an integer that is the maximum number of records that can be returned in one page.
	//
	// ---
	// responses:
	//   '200':
	//     description: The page max records limit was returned successfully.
	r.HandleFunc(productsPageMaxRecordLimitAPIRoute, s.getPageMaxRecordLimitAPIHandler()).Methods(s.getPageMaxRecordLimitAPIOptions().AllowedMethods...)
	// swagger:operation GET /products/health products checkHealth
	//
	// Checks the health of the service.
	//
	// ---
	// responses:
	//   '200':
	//     description: The health check was completed.
	//     "$ref": "#/responses/healthCheckResponse"
	r.HandleFunc(productsHealthAPIRoute, s.getHealthCheckAPIHandler()).Methods(s.getHealthCheckAPIOptions().AllowedMethods...)

	return r
}

// CheckHealth checks the health of the product listing service and writes a response in JSON to the user.
// Always returns HTTP Status OK, even if the health check fails.
func (s *ProductsService) CheckHealth(w http.ResponseWriter, r *http.Request) {
	var err error
	log.Info("Checking products service health...")
	db, err := s.DB.Postgres.DB()
	if err != nil {
		log.Error("products service health check failed: Error getting SQLDB from gorm DB: " + err.Error())
		json.WriteResponse(w, http.StatusOK, map[string]bool{"ok": false})
	} else {
		if err = db.Ping(); err != nil {
			log.Error("products service health check failed: error pinging the database: " + err.Error())
			json.WriteResponse(w, http.StatusOK, map[string]bool{"ok": false})
		} else {
			log.Info("products service health check passed")
			json.WriteResponse(w, http.StatusOK, map[string]bool{"ok": true})
		}
	}
}

// SetupOrdersServiceDB checks that the database schema is ready for the authentication service.
// If init is true, will create the tables if they do not already exist.
func setupProductsServiceDB(db *driver.DB, init bool) error {
	log.Info("Setting up the products service database...")
	err := pgdriver.SetupTables(db, &models.Product{}, init)
	if err != nil {
		msg := "failed to set up the Product model table" + err.Error()
		log.Error(msg)
		return errors.New(msg)
	}
	log.Info("Successfully set up the database for the products service")
	return nil
}
