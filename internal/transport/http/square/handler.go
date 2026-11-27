package square

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/neh333/KaininAPIV2/internal/model"
	"github.com/neh333/KaininAPIV2/internal/service"
	"github.com/sirupsen/logrus"
	squareSDK "github.com/square/square-go-sdk/v5"
	squareclient "github.com/square/square-go-sdk/v5/client"
	squarecore "github.com/square/square-go-sdk/v5/core"
	squarewebhookclient "github.com/square/square-go-sdk/v5/webhooks/client"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"gorm.io/gorm"
)

const (
	webhookPath           = "/api/square/webhook"
	shippingAddressPath   = "/api/account/shipping-address"
	maxWebhookBody        = 2 << 20
	maxAddressRequestBody = 16 << 10
)

type Handler struct {
	service         *service.SquareService
	db              *gorm.DB
	square          *squareclient.Client
	signatureKey    string
	notificationURL string
	verifier        *squarewebhookclient.Client
}

func RegisterRoutes(router chi.Router, db *gorm.DB, signatureKey, notificationURL string) {
	handler := &Handler{
		service:         service.NewSquareService(db),
		db:              db,
		square:          squareclient.NewClient(),
		signatureKey:    strings.TrimSpace(signatureKey),
		notificationURL: strings.TrimSpace(notificationURL),
		verifier:        squarewebhookclient.NewClient(squarecore.NewRequestOptions()),
	}
	router.Post(webhookPath, handler.ReceiveWebhook)
	router.Get(shippingAddressPath, handler.GetShippingAddress)
	router.Put(shippingAddressPath, handler.SaveShippingAddress)
}

type shippingAddress struct {
	AddressLine1                 string `json:"addressLine1"`
	AddressLine2                 string `json:"addressLine2"`
	Locality                     string `json:"locality"`
	AdministrativeDistrictLevel1 string `json:"administrativeDistrictLevel1"`
	PostalCode                   string `json:"postalCode"`
	Country                      string `json:"country"`
}

type shippingAddressResult struct {
	Address *shippingAddress `json:"address"`
}

func (h *Handler) GetShippingAddress(w http.ResponseWriter, r *http.Request) {
	if !h.ensureSquareConfigured(w) {
		return
	}
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}

	customer, err := h.findSquareCustomer(r, user)
	if err != nil {
		logrus.WithError(err).Error("Failed to load the user's Square customer")
		http.Error(w, "Unable to load shipping address", http.StatusBadGateway)
		return
	}

	var address *shippingAddress
	if customer != nil && customer.Address != nil {
		address = addressFromSquare(customer.Address)
	}
	writeJSON(w, http.StatusOK, shippingAddressResult{Address: address})
}

func (h *Handler) SaveShippingAddress(w http.ResponseWriter, r *http.Request) {
	if !h.ensureSquareConfigured(w) {
		return
	}
	user, ok := h.currentUser(w, r)
	if !ok {
		return
	}

	var address shippingAddress
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxAddressRequestBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&address); err != nil {
		http.Error(w, "Invalid shipping address", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "Request must contain one shipping address", http.StatusBadRequest)
		return
	}
	if err := validateShippingAddress(&address); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	squareAddress := address.toSquareAddress()
	customer, err := h.findSquareCustomer(r, user)
	if err != nil {
		logrus.WithError(err).Error("Failed to find the user's Square customer")
		http.Error(w, "Unable to save shipping address", http.StatusBadGateway)
		return
	}

	if customer == nil {
		referenceID := strconv.FormatInt(user.UserID, 10)
		idempotencyKey := fmt.Sprintf("kainin-user-%d", user.UserID)
		response, err := h.square.Customers.Create(r.Context(), &squareSDK.CreateCustomerRequest{
			IdempotencyKey: &idempotencyKey,
			GivenName:      optionalString(user.FirstName),
			FamilyName:     optionalString(user.LastName),
			EmailAddress:   optionalString(user.Email),
			ReferenceID:    &referenceID,
			Address:        squareAddress,
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to create a Square customer for the user")
			http.Error(w, "Unable to save shipping address", http.StatusBadGateway)
			return
		}
		customer = response.GetCustomer()
		if customer == nil || customer.GetID() == nil {
			http.Error(w, "Square did not return the saved customer", http.StatusBadGateway)
			return
		}
		if err := h.persistSquareCustomerID(r, user.UserID, *customer.ID); err != nil {
			logrus.WithError(err).Error("Failed to store the user's Square customer ID")
			http.Error(w, "Shipping address was saved but could not be linked to the account", http.StatusInternalServerError)
			return
		}
	} else {
		if customer.GetID() == nil {
			http.Error(w, "Square customer is missing its ID", http.StatusBadGateway)
			return
		}
		request := &squareSDK.UpdateCustomerRequest{
			CustomerID: *customer.ID,
			Address:    squareAddress,
			Version:    customer.Version,
		}
		response, err := h.square.Customers.Update(r.Context(), request)
		if err != nil {
			logrus.WithError(err).Error("Failed to update the user's Square customer")
			http.Error(w, "Unable to save shipping address", http.StatusBadGateway)
			return
		}
		customer = response.GetCustomer()
		if customer == nil {
			http.Error(w, "Square did not return the saved customer", http.StatusBadGateway)
			return
		}
	}

	writeJSON(w, http.StatusOK, shippingAddressResult{Address: addressFromSquare(customer.Address)})
}

func (h *Handler) ensureSquareConfigured(w http.ResponseWriter) bool {
	if strings.TrimSpace(os.Getenv("SQUARE_TOKEN")) == "" {
		http.Error(w, "Square customer access is not configured", http.StatusServiceUnavailable)
		return false
	}
	return true
}

func (h *Handler) currentUser(w http.ResponseWriter, r *http.Request) (model.User, bool) {
	sessionContainer := session.GetSessionFromRequestContext(r.Context())
	if sessionContainer == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return model.User{}, false
	}

	var user model.User
	err := h.db.WithContext(r.Context()).
		Where("supertokens_user_id = ?", sessionContainer.GetUserID()).
		Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "No application user is linked to this session", http.StatusForbidden)
		return model.User{}, false
	}
	if err != nil {
		logrus.WithError(err).Error("Failed to load the authenticated application user")
		http.Error(w, "Unable to load user", http.StatusInternalServerError)
		return model.User{}, false
	}
	return user, true
}

func (h *Handler) findSquareCustomer(r *http.Request, user model.User) (*squareSDK.Customer, error) {
	if user.SquareCustomerID != nil && *user.SquareCustomerID != "" {
		response, err := h.square.Customers.Get(r.Context(), &squareSDK.GetCustomersRequest{
			CustomerID: *user.SquareCustomerID,
		})
		if err != nil {
			return nil, err
		}
		if customer := response.GetCustomer(); customer != nil {
			return customer, nil
		}
		return nil, fmt.Errorf("Square returned an empty customer response")
	}

	referenceID := strconv.FormatInt(user.UserID, 10)
	limit := int64(1)
	response, err := h.square.Customers.Search(r.Context(), &squareSDK.SearchCustomersRequest{
		Limit: &limit,
		Query: &squareSDK.CustomerQuery{
			Filter: &squareSDK.CustomerFilter{
				ReferenceID: &squareSDK.CustomerTextFilter{Exact: &referenceID},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("Square returned an empty customer search response")
	}
	customers := response.GetCustomers()
	if len(customers) == 0 {
		return nil, nil
	}
	customer := customers[0]
	if customer == nil || customer.GetID() == nil {
		return nil, fmt.Errorf("Square returned a customer without an ID")
	}
	if err := h.persistSquareCustomerID(r, user.UserID, *customer.ID); err != nil {
		return nil, err
	}
	return customer, nil
}

func (h *Handler) persistSquareCustomerID(r *http.Request, userID int64, customerID string) error {
	return h.db.WithContext(r.Context()).
		Model(&model.User{}).
		Where("UserID = ?", userID).
		Update("SquareCustomerID", customerID).Error
}

func (a *shippingAddress) toSquareAddress() *squareSDK.Address {
	country := squareSDK.Country(a.Country)
	return &squareSDK.Address{
		AddressLine1:                 optionalString(a.AddressLine1),
		AddressLine2:                 optionalString(a.AddressLine2),
		Locality:                     optionalString(a.Locality),
		AdministrativeDistrictLevel1: optionalString(a.AdministrativeDistrictLevel1),
		PostalCode:                   optionalString(a.PostalCode),
		Country:                      &country,
	}
}

func addressFromSquare(address *squareSDK.Address) *shippingAddress {
	if address == nil {
		return nil
	}
	country := ""
	if value := address.GetCountry(); value != nil {
		country = string(*value)
	}
	return &shippingAddress{
		AddressLine1:                 stringValue(address.GetAddressLine1()),
		AddressLine2:                 stringValue(address.GetAddressLine2()),
		Locality:                     stringValue(address.GetLocality()),
		AdministrativeDistrictLevel1: stringValue(address.GetAdministrativeDistrictLevel1()),
		PostalCode:                   stringValue(address.GetPostalCode()),
		Country:                      country,
	}
}

func validateShippingAddress(address *shippingAddress) error {
	address.AddressLine1 = strings.TrimSpace(address.AddressLine1)
	address.AddressLine2 = strings.TrimSpace(address.AddressLine2)
	address.Locality = strings.TrimSpace(address.Locality)
	address.AdministrativeDistrictLevel1 = strings.TrimSpace(address.AdministrativeDistrictLevel1)
	address.PostalCode = strings.TrimSpace(address.PostalCode)
	address.Country = strings.ToUpper(strings.TrimSpace(address.Country))
	if address.AddressLine1 == "" || address.Locality == "" || address.PostalCode == "" {
		return fmt.Errorf("street address, city, and postal code are required")
	}
	if len(address.Country) != 2 || address.Country[0] < 'A' || address.Country[0] > 'Z' ||
		address.Country[1] < 'A' || address.Country[1] > 'Z' {
		return fmt.Errorf("country must be a two-letter ISO country code")
	}
	return nil
}

func optionalString(value string) *string {
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		logrus.WithError(err).Error("Failed to write JSON response")
	}
}

func (h *Handler) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	if h.signatureKey == "" || h.notificationURL == "" {
		http.Error(w, "Square webhook verification is not configured", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBody+1))
	if err != nil {
		http.Error(w, "Unable to read webhook request", http.StatusBadRequest)
		return
	}
	if len(body) > maxWebhookBody {
		http.Error(w, "Webhook request is too large", http.StatusRequestEntityTooLarge)
		return
	}
	if len(body) == 0 {
		http.Error(w, "Webhook request body is required", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("x-square-hmacsha256-signature")
	err = h.verifier.VerifySignature(
		r.Context(),
		&squareSDK.VerifySignatureRequest{
			RequestBody:     string(body),
			SignatureHeader: signature,
			SignatureKey:    h.signatureKey,
			NotificationURL: h.notificationURL,
		},
	)
	if err != nil {
		logrus.WithError(err).Warn("Rejected Square webhook with invalid signature")
		http.Error(w, "Invalid Square webhook signature", http.StatusForbidden)
		return
	}

	var event service.SquareWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid Square webhook payload", http.StatusBadRequest)
		return
	}
	if err := h.service.ProcessWebhook(r.Context(), event); err != nil {
		logrus.WithError(err).Error("Failed to process Square webhook")
		http.Error(w, "Failed to process Square webhook", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
