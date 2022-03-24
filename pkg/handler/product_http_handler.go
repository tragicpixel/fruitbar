package handler

import (
	"strings"

	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/models/roles"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	itemsrepo "github.com/tragicpixel/fruitbar/pkg/repository/item"
	jwtrepo "github.com/tragicpixel/fruitbar/pkg/repository/jwt"
	productrepo "github.com/tragicpixel/fruitbar/pkg/repository/product"
	"github.com/tragicpixel/fruitbar/pkg/utils"
	jwtutils "github.com/tragicpixel/fruitbar/pkg/utils/jwt"
	"github.com/tragicpixel/fruitbar/pkg/utils/log"

	"fmt"
	"net/http"
)

// Product represents a handler for performing operations on products via HTTP.
type Product struct {
	repo      repository.Product
	itemsRepo repository.Item
	jwtRepo   repository.Jwt
}

// NewProductHandler creates and initializes a new handler for performing operations on products via HTTP.
func NewProductHandler(db *driver.DB) *Product {
	return &Product{
		repo:      productrepo.NewPostgresProductRepo(db.Postgres),
		itemsRepo: itemsrepo.NewPostgresItemRepo(db.Postgres),
		jwtRepo:   jwtrepo.NewJWTRepository(),
	}
}

// CreateProduct creates a new product based on the supplied HTTP request and sends a response in JSON containing the newly created product to the supplied http response writer.
// If there is a permission error, an HTTP error will be sent.
func (h *Product) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if !h.clientHasCreatePerms(w, r) {
		return
	}

	var product models.Product
	response := *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &product, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}
	_, err := product.IsValid()
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Product "+validationFailedErrMsgPrefix+err.Error())
		return
	}

	log.Info(fmt.Sprintf("Inserting new Product: %+v", product))
	createdId, err := h.repo.Create(&product)
	if err != nil {
		logMsg := fmt.Sprintf("Error inserting Product: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	log.Info(fmt.Sprintf("Created new product (id: %d): %+v", createdId, product))
	response = utils.JsonResponse{Data: []*models.Product{&product}}
	utils.WriteJSONResponse(w, http.StatusCreated, response)
}

// GetProducts sends a response to the supplied http response writer containing the requested product(s), based on the supplied http request.
func (h *Product) GetProducts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has(idParam) {
		h.getSingleProduct(w, r)
	} else {
		h.getProductsPage(w, r)
	}
}

// UpdateOrder updates an existing product based on the supplied http request and sends a response in JSON containing the updated product to the supplied http response writer.
func (h *Product) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if !h.clientHasUpdatePerms(w, r) {
		return
	}

	var product models.Product
	response := *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &product, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	if r.URL.Query().Has(fieldsParam) {
		h.partiallyUpdateProduct(w, r, product)
	} else {
		h.fullyUpdateProduct(w, r, product)
	}
}

// DeleteOrder deletes an existing product and any items with the its product ID, based on the supplied http request, and sends a status code to the supplied http response writer.
func (h *Product) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if !h.clientHasDeletePerms(w, r) {
		return
	}

	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Info(fmt.Sprintf("Selecting items with product id %d for potential delete...", id))
	existingItems, err := h.itemsRepo.GetByProductID(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading existing items for product (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	for _, item := range existingItems {
		log.Info(fmt.Sprintf("Deleting existing item (id: %d) from product (id: %d)", item.ID, id))
		err := h.itemsRepo.Delete(item.ID)
		if err != nil {
			logMsg := fmt.Sprintf("Error deleting existing item (id: %d): %s", item.ID, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
	}
	log.Info(fmt.Sprintf("Deleted all items for product (id: %d)", id))

	log.Info(fmt.Sprintf("Deleting product (id: %d)...", id))
	err = h.repo.Delete(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error deleting product (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	log.Info(fmt.Sprintf("Successfully deleted product with id = %d.", id))
	w.WriteHeader(http.StatusNoContent)
}

// getSingleOrder sends a response to the supplied http response writer containing the requested product, based on the supplied http request.
func (h *Product) getSingleProduct(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Info(fmt.Sprintf("Reading product (id: %d)...", id))
	var product *models.Product
	product, err = h.repo.GetByID(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading product (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	log.Info(fmt.Sprintf("Read product (id: %d)", id))
	response := utils.JsonResponse{Data: []*models.Product{product}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// getProductsPage sends a response to the supplied http response writer containing the requested page of products, based on the supplied http request.
func (h *Product) getProductsPage(w http.ResponseWriter, r *http.Request) {
	seek, err := utils.GetPageSeekOptions(r, readPageMaxLimit)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Info(fmt.Sprintf("Reading %d products (max %d)...", seek.RecordLimit, readPageMaxLimit))
	var products []*models.Product
	products, err = h.repo.Fetch(seek)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading products: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}

	rangeStr := h.getProductsRangeStr(w, seek, products)
	w.Header().Set("Content-Range", rangeStr)
	log.Info(fmt.Sprintf("Read %d products", len(products)))
	response := utils.JsonResponse{Data: products}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// partiallyUpdateProduct updates only the specified fields (via http query parameter) of the supplied user
// and sends a response to the supplied http response writer containing the updated user in JSON.
func (h *Product) partiallyUpdateProduct(w http.ResponseWriter, r *http.Request, product models.Product) {
	fieldsStr := r.URL.Query().Get(fieldsParam)
	fields := strings.Split(fieldsStr, ",")

	_, err := product.PartialUpdateIsValid(fields)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Product "+validationFailedErrMsgPrefix+err.Error())
		return
	}

	log.Info(fmt.Sprintf("Updating Product (id: %d) fields (%s) to %+v", product.ID, fieldsStr, product))
	updated, err := h.repo.Update(&product, fields)
	if err != nil {
		logMsg := fmt.Sprintf("Error partially updating Product (id: %d)  fields (%s) : %s", product.ID, fieldsStr, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	log.Info(fmt.Sprintf("Partially updated Product (id: %d) fields (%s): %+v", product.ID, fieldsStr, updated))
	response := utils.JsonResponse{Data: []*models.Product{&product}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// fullyUpdateProduct updates all of the fields of the supplied product
// and sends a response to the supplied http response writer containing the updated product in JSON.
func (h *Product) fullyUpdateProduct(w http.ResponseWriter, r *http.Request, product models.Product) {
	_, err := product.IsValid()
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Product "+validationFailedErrMsgPrefix+err.Error())
		return
	}

	log.Info(fmt.Sprintf("Updating Product (id: %d) to %+v", product.ID, product))
	updated, err := h.repo.Update(&product, []string{})
	if err != nil {
		logMsg := fmt.Sprintf("Error fully updating Product with id = %d: %+v: %s", product.ID, product, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	log.Info(fmt.Sprintf("Fully updated Product (id: %d): %+v", product.ID, updated))
	response := utils.JsonResponse{Data: []*models.Product{&product}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// getClientAuthInfo returns the authorization information about the client based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *Product) getClientAuthInfo(w http.ResponseWriter, r *http.Request) *models.JwtClaim {
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

// clientHasCreatePerms checks whether the client has permissions to create a product, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *Product) clientHasCreatePerms(w http.ResponseWriter, r *http.Request) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Only an admin can create a product
	if client.UserRole == roles.Admin {
		http.Error(w, forbiddenCreateProductErrMsg, http.StatusForbidden)
		return false
	}
	return true
}

// clientHasUpdatePerms checks whether the client has permissions to update a product, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *Product) clientHasUpdatePerms(w http.ResponseWriter, r *http.Request) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Only an admin can update a product
	if client.UserRole == roles.Admin {
		http.Error(w, forbiddenUpdateProductErrMsg, http.StatusForbidden)
		return false
	}
	return true
}

// clientHasDeletePerms checks whether the client has permissions to delete a product, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *Product) clientHasDeletePerms(w http.ResponseWriter, r *http.Request) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Only an admin can delete a product
	if client.UserRole == roles.Admin {
		http.Error(w, forbiddenDeleteProductErrMsg, http.StatusForbidden)
		return false
	}
	return true
}

// getProductsRangeStr returns a string representation of the range of the supplied products.
func (h *Product) getProductsRangeStr(w http.ResponseWriter, seek *repository.PageSeekOptions, products []*models.Product) string {
	log.Info("Counting products...")
	// TODO: Cache this count value and update every X seconds, so we don't need to perform a full count on every page read.
	// TODO: I want a full count here, but I think this is just returning the number of total records based on this seek, not the total # of orders.
	count, err := h.repo.Count(seek)
	if err != nil {
		logMsg := fmt.Sprintf("Error counting products: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return ""
	}
	startID, endID := uint(0), uint(0)
	if len(products) > 0 {
		startID = products[0].ID
		endID = products[len(products)-1].ID
	}
	rangeStr := fmt.Sprintf("products=%d-%d/%d", startID, endID, count)
	return rangeStr
}
