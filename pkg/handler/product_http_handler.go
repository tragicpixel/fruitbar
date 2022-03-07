package handler

import (
	"strings"

	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	productrepo "github.com/tragicpixel/fruitbar/pkg/repository/product"
	"github.com/tragicpixel/fruitbar/pkg/utils"

	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Product represents an http handler for performing operations on a repository of products.
type Product struct {
	repo repository.Product
}

// NewProductHandler creates a new http handler for performing operations on a repository of products.
func NewProductHandler(db *driver.DB) *Product {
	return &Product{
		repo: productrepo.NewPostgresProductRepo(db.Postgres), // this is where it is decided which implementation(/database type) of the Product Repo we will use
	}
}

// CreateOrder creates a new product in the repo based on the supplied HTTP request and sends a response in JSON to the user based on success or failure.
// Requires a name, a symbol consisting of 1 emoji, a price > 0, and a number in stock >= 0.
// If number in stock isn't specified, will default to zero.
func (h *Product) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// TODO: remove this and rewrite the other code to use the CORS middleware
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodPost, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Create Product API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}
	var product models.Product
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &product, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error == nil { // New product json successfully decoded
		// Validate new product
		newProductIsValid, newProductValidationError := models.ValidateProduct(&product) // TODO: Write ValidateNewProduct method to validate new products instead
		if !newProductIsValid {
			msg := "Failed to validate new product: " + newProductValidationError.Error()
			logrus.Error(msg)
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: msg}}
		} else { // New product is valid
			// Create the new product
			logrus.Info(fmt.Sprintf("Trying to insert new Product: %+v", product))
			createdId, err := h.repo.Create(&product)
			if err != nil {
				logrus.Error(fmt.Sprintf("Error inserting Product %+v into database: %s", product, err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create new product."}}
			} else {
				logrus.Info(fmt.Sprintf("Successfully inserted new product: %+v", product))
				response = utils.JsonResponse{Data: []*models.Product{&product}, Id: strconv.Itoa(int(createdId))}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else { // Create was successful
		utils.WriteJSONResponse(w, http.StatusCreated, response)
	}
}

func (h *Product) getReadSingleProductIdParamName() string { return "id" }
func (h *Product) getReadPageMaxLimit() int                { return 2 }

// GetProducts retrieves either a single product in the database if the "id" query parameter is supplied, or a list of products in the database if it is not.
// The list of products can also be paginated, using the "limit" and "after_id" query parameters.
// Returns a response in JSON containing either an array of products encoded in JSON or an error message.
func (h *Product) GetProducts(w http.ResponseWriter, r *http.Request) {
	// TODO: remove this and rewrite the other code to use the CORS middleware
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodGet, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Read Products API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}
	if !r.URL.Query().Has(h.getReadSingleProductIdParamName()) { // Single product id parameter is not set
		// Get page seek options
		var pageSeekOptions utils.PageSeekOptions
		pageSeekOptions, response = utils.GetPageSeekOptions(r, h.getReadPageMaxLimit())
		if response.Error == nil { // Page seek options are valid
			// Read the multiple products
			logrus.Info("Retrieving " + strconv.Itoa(pageSeekOptions.RecordLimit) + " products (max " + strconv.Itoa(h.getReadPageMaxLimit()) + ")...")
			var products []*models.Product
			products, err := h.repo.Fetch(pageSeekOptions)

			if err != nil {
				logrus.Error(fmt.Sprintf("Error retrieving products list: %s", err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "Failed to retrieve products list"}}
			} else {
				logrus.Info("Successfully retrieved products list")
				response = utils.JsonResponse{Data: products}
			}
		}
	} else { // Single product ID parameter is set
		// Validate product id query parameter
		id, err := utils.GetQueryParamAsInt(r, h.getReadSingleProductIdParamName())
		if err != nil {
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}}
		} else { // Product ID is valid
			// Read the single product
			logrus.Info("Retrieving product with ID = " + strconv.Itoa(id) + "...")
			var product *models.Product
			product, err := h.repo.GetByID(id)
			if err != nil {
				logrus.Error(fmt.Sprintf("Error retrieving product with id = %d: %s", id, err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "Failed to retrieve product with id = " + strconv.Itoa(id)}}
			} else {
				logrus.Info("Successfully retrieved product with id = " + strconv.Itoa(id))
				response = utils.JsonResponse{Data: []*models.Product{product}}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else { // Read was successful
		utils.WriteJSONResponse(w, http.StatusOK, response)
	}
}

const (
	fieldsParamName = "fields"
)

// UpdateProduct updates an existing product in the repo based on the supplied JSON request, and returns a status message in JSON to the user.
// If price is not set or set to zero, it will be ignored.
func (h *Product) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// TODO: remove this and rewrite the other code to use the CORS middleware
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodPut, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Update Products API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}
	var product models.Product
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &product, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error == nil { // Updated product was successfully decoded
		if r.URL.Query().Has(fieldsParamName) { // Partial update of order was requested
			// Validate the partial update's changes
			fields := strings.Split(r.URL.Query().Get(fieldsParamName), ",")
			isProductUpdateValid, err := models.ValidatePartialProductUpdate(&product, fields)
			if !isProductUpdateValid {
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}}
			} else { // Partial update is valid
				// Partially update the product
				logrus.Info(fmt.Sprintf("Partially updating Product with id = %d to %+v", product.ID, product))
				_, err := h.repo.Update(&product, fields)
				if err != nil {
					logrus.Error(fmt.Sprintf("Error partially updating Product with id = %d: %+v: %s", product.ID, product, err.Error()))
					response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "Internal error: Failed to partially update existing product.")
				} else {
					logrus.Info(fmt.Sprintf("Successfully partially updated Product with id = %d: %+v", product.ID, product))
					response = utils.JsonResponse{Data: []*models.Product{&product}, Id: strconv.Itoa(int(product.ID))}
				}
			}
		} else { // Full update of order was requested
			isValid, err := models.ValidateProduct(&product)
			if !isValid {
				response = utils.NewJsonResponseWithError(http.StatusBadRequest, err.Error())
			} else { // Full update is valid
				// Fully update the product
				logrus.Info(fmt.Sprintf("Partially updating Product with id = %d to %+v", product.ID, product))
				_, err := h.repo.Update(&product, []string{})
				if err != nil {
					logrus.Error(fmt.Sprintf("Error fully updating Product with id = %d: %+v: %s", product.ID, product, err.Error()))
					response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "Internal error: Failed to fully update existing product.")
				} else {
					logrus.Info(fmt.Sprintf("Successfully fully updated Product with id = %d: %+v", product.ID, product))
					response = utils.JsonResponse{Data: []*models.Product{&product}, Id: strconv.Itoa(int(product.ID))}
				}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else { // Update was successful
		utils.WriteJSONResponse(w, http.StatusOK, response)
	}
}

func (h *Product) getDeleteProductIdParamName() string { return "id" }

// DeleteProduct deletes an existing product from the repo based on the supplied http request, and returns a status message in JSON to the user.
func (h *Product) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// TODO: remove this and rewrite the other code to use the CORS middleware
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodDelete, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Delete Products API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}

	// Validate product id parameter
	id, err := utils.GetQueryParamAsInt(r, h.getDeleteProductIdParamName())
	if err != nil {
		response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}}
	}

	if response.Error == nil { // Product ID is valid
		// Delete the product
		logrus.Info("Deleting product with id " + strconv.Itoa(id) + "...")
		_, err := h.repo.Delete(int64(id))
		if err != nil {
			logrus.Error(fmt.Sprintf("Error deleting product with id = %d: %s", id, err.Error()))
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete product with id = " + strconv.Itoa(id)}}
		} else {
			logrus.Info(fmt.Sprintf("Successfully deleted product with id = %d.", id))
		}

		// Also delete any items using this product ID
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else { // Delete was successful
		utils.WriteJSONResponse(w, http.StatusNoContent, response)
	}
}
