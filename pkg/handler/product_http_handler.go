package handler

import (
	"strings"

	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	itemsrepo "github.com/tragicpixel/fruitbar/pkg/repository/item"
	productrepo "github.com/tragicpixel/fruitbar/pkg/repository/product"
	"github.com/tragicpixel/fruitbar/pkg/utils"

	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Product represents an http handler for performing operations on a repository of products.
type Product struct {
	repo      repository.Product
	itemsRepo repository.Item
}

// NewProductHandler creates a new http handler for performing operations on a repository of products.
func NewProductHandler(db *driver.DB) *Product {
	return &Product{
		repo:      productrepo.NewPostgresProductRepo(db.Postgres),
		itemsRepo: itemsrepo.NewPostgresItemRepo(db.Postgres),
	}
}

// CreateProduct creates a new product in the repo based on the supplied HTTP request and writes a response over HTTP.
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
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	_, err := models.ValidateProduct(&product) // TODO: Write ValidateNewProduct method to validate new products instead
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Product "+validationFailedMsg+err.Error())
		return
	}

	logrus.Info(fmt.Sprintf("Creating new Product: %+v", product))
	createdId, err := h.repo.Create(&product)
	if err != nil {
		logMsg := fmt.Sprintf("Error creating Product: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Created new product (id: %d): %+v", createdId, product))
	//product.ID = createdId
	response = utils.JsonResponse{Data: []*models.Product{&product}}
	utils.WriteJSONResponse(w, http.StatusCreated, response)
}

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
	if r.URL.Query().Has(idParam) {
		// Read a single product
		id, err := utils.GetQueryParamAsUint(r, idParam)
		if err != nil {
			utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		logrus.Info(fmt.Sprintf("Reading product (id: %d)...", id))
		var product *models.Product
		product, err = h.repo.GetByID(id)
		if err != nil {
			logMsg := fmt.Sprintf("Error reading product (id: %d): %s", id, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
		logrus.Info(fmt.Sprintf("Read product (id: %d)", id))
		response = utils.JsonResponse{Data: []*models.Product{product}}
		utils.WriteJSONResponse(w, http.StatusOK, response)
	} else {
		// Read a page of products
		var seek *utils.PageSeekOptions
		seek, err := utils.GetPageSeekOptions(r, readPageMaxLimit)
		if response.Error != nil {
			utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		logrus.Info(fmt.Sprintf("Reading %d products (max %d)...", seek.RecordLimit, readPageMaxLimit))
		var products []*models.Product
		products, err = h.repo.Fetch(seek)
		if err != nil {
			logMsg := fmt.Sprintf("Error reading products: %s", err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
		count, err := h.repo.Count(seek)
		if err != nil {
			logMsg := fmt.Sprintf("Error counting products: %s", err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
		startID, endID := uint(0), uint(0)
		if len(products) > 0 {
			startID = products[0].ID
			endID = products[len(products)-1].ID
		}
		rangeStr := fmt.Sprintf("Range: products=%d-%d/%d", startID, endID, count)
		w.Header().Set("Range", rangeStr)
		logrus.Info(fmt.Sprintf("Read %d products", len(products)))
		response = utils.JsonResponse{Data: products}
		utils.WriteJSONResponse(w, http.StatusOK, response)
	}
}

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
	if response.Error != nil { // Updated product was successfully decoded
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
	}

	if r.URL.Query().Has(fieldsParam) {
		// Partially update product
		fieldsStr := r.URL.Query().Get(fieldsParam)
		fields := strings.Split(fieldsStr, ",")

		_, err := models.ValidatePartialProductUpdate(&product, fields)
		if err != nil {
			utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Product "+validationFailedMsg+err.Error())
			return
		}

		logrus.Info(fmt.Sprintf("Updating Product (id: %d) fields (%s) to %+v", product.ID, fieldsStr, product))
		updated, err := h.repo.Update(&product, fields)
		if err != nil {
			logMsg := fmt.Sprintf("Error partially updating Product (id: %d)  fields (%s) : %s", product.ID, fieldsStr, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
		logrus.Info(fmt.Sprintf("Partially updated Product (id: %d) fields (%s): %+v", product.ID, fieldsStr, updated))
		response = utils.JsonResponse{Data: []*models.Product{&product}, Id: strconv.Itoa(int(product.ID))}
		utils.WriteJSONResponse(w, http.StatusOK, response)
		return
	} else {
		// Fully update product
		_, err := models.ValidateProduct(&product)
		if err != nil {
			utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Product "+validationFailedMsg+err.Error())
			return
		}

		logrus.Info(fmt.Sprintf("Updating Product (id: %d) to %+v", product.ID, product))
		updated, err := h.repo.Update(&product, []string{})
		if err != nil {
			logrus.Error(fmt.Sprintf("Error fully updating Product with id = %d: %+v: %s", product.ID, product, err.Error()))
			response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "Internal error: Failed to fully update existing product.")
			return
		}
		logrus.Info(fmt.Sprintf("Fully updated Product (id: %d): %+v", product.ID, updated))
		response = utils.JsonResponse{Data: []*models.Product{&product}}
		utils.WriteJSONResponse(w, http.StatusOK, response)
	}
}

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

	// Validate product id parameter
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Delete any items with this product ID
	existingItems, err := h.itemsRepo.GetByProductID(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading existing items for product (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	for _, item := range existingItems {
		logrus.Info(fmt.Sprintf("Deleting existing item (id: %d) from product (id: %d)", item.ID, id))
		err := h.itemsRepo.Delete(item.ID)
		if err != nil {
			logMsg := fmt.Sprintf("Error deleting existing item (id: %d): %s", item.ID, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
	}
	logrus.Info(fmt.Sprintf("Deleted all items for product (id: %d)", id))

	// Delete the product
	logrus.Info(fmt.Sprintf("Deleting product (id: %d)...", id))
	err = h.repo.Delete(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error deleting product (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Successfully deleted product with id = %d.", id))
	response := utils.JsonResponse{}
	utils.WriteJSONResponse(w, http.StatusNoContent, response)
}
