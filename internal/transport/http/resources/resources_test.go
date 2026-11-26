package resources

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/neh333/KaininAPIV2/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestResponseDTOsOmitAuthenticationSecrets(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		secret string
	}{
		{
			name: "user",
			value: userResponse(model.User{
				Email:             "user@example.com",
				SupertokensUserID: "super-secret-user-id",
			}),
			secret: "super-secret-user-id",
		},
		{
			name: "login method",
			value: loginMethodResponse(model.LoginMethod{
				SupertokensUserID: "super-secret-user-id",
				ProviderUserID:    "private-provider-id",
				Provider:          "google",
			}),
			secret: "private-provider-id",
		},
		{
			name: "newsletter subscription",
			value: newsletterSubscriptionResponse(model.NewsletterSubscription{
				Email: "user@example.com",
				Token: "verification-secret",
			}),
			secret: "verification-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), tt.secret) {
				t.Fatalf("response leaked a sensitive value: %s", encoded)
			}
		})
	}
}

func TestRegisterRoutesBuildsOpenAPISpec(t *testing.T) {
	api := humachi.New(chi.NewRouter(), huma.DefaultConfig("test", "1.0.0"))
	RegisterRoutes(api, &gorm.DB{})

	paths := api.OpenAPI().Paths
	for _, path := range []string{
		"/api/bags",
		"/api/bag-items",
		"/api/products",
		"/api/product-variations",
		"/api/newsletter-subscriptions",
		"/api/users/me",
	} {
		if _, ok := paths[path]; !ok {
			t.Errorf("OpenAPI path %q was not registered", path)
		}
	}
	if _, ok := paths["/api/login-methods"]; ok {
		t.Error("authentication internals must not have a public API route")
	}
}

func TestGORMModelsHavePrimaryKeysForCRUD(t *testing.T) {
	tests := []struct {
		name  string
		value any
		field string
	}{
		{name: "bag item", value: &model.BagItem{}, field: "BagItemID"},
		{name: "bag", value: &model.Bag{}, field: "BagID"},
		{name: "login method", value: &model.LoginMethod{}, field: "id"},
		{name: "newsletter subscription", value: &model.NewsletterSubscription{}, field: "Id"},
		{name: "option type", value: &model.OptionType{}, field: "OptionTypeID"},
		{name: "option value", value: &model.OptionValue{}, field: "OptionValueID"},
		{name: "order item", value: &model.OrderItem{}, field: "OrderItemID"},
		{name: "order", value: &model.Order{}, field: "OrderID"},
		{name: "order status", value: &model.OrderStatus{}, field: "OrderStatusID"},
		{name: "product option", value: &model.ProductOption{}, field: "ProductOptionID"},
		{name: "product", value: &model.Product{}, field: "ProductID"},
		{name: "product variation", value: &model.ProductVariation{}, field: "ProductVariationID"},
		{name: "role", value: &model.Role{}, field: "RoleID"},
		{name: "user", value: &model.User{}, field: "UserID"},
	}

	var cache sync.Map
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := schema.Parse(tt.value, &cache, schema.NamingStrategy{})
			if err != nil {
				t.Fatal(err)
			}
			if parsed.PrioritizedPrimaryField == nil || parsed.PrioritizedPrimaryField.DBName != tt.field {
				t.Fatalf("primary key = %v, want %q", parsed.PrioritizedPrimaryField, tt.field)
			}
		})
	}
}

func TestProductUpdateMapperIncludesZeroValues(t *testing.T) {
	fields := productUpdates(ProductRequest{})
	if fields["Price"] != float64(0) || fields["Name"] != "" || fields["image_uri"] != "" {
		t.Fatalf("update map omitted zero values: %#v", fields)
	}
	if _, ok := fields["ProductID"]; ok {
		t.Fatalf("update map should not permit changing a product ID: %#v", fields)
	}
}

func TestUserPatchMapperUpdatesOnlyProvidedFields(t *testing.T) {
	empty := ""
	fields := userUpdates(UserRequest{PhoneNumber: &empty})
	if len(fields) != 1 || fields["PhoneNumber"] != "" {
		t.Fatalf("expected explicit empty phone number only, got %#v", fields)
	}
}

func TestCatalogPublicAccessIsReadOnly(t *testing.T) {
	tests := []struct {
		method string
		path   string
		public bool
	}{
		{method: "GET", path: "/api/products", public: true},
		{method: "GET", path: "/api/products/7", public: true},
		{method: "GET", path: "/api/orders", public: false},
		{method: "GET", path: "/api/users/me", public: false},
		{method: "POST", path: "/api/products", public: false},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			if got := isPublicCatalogRead(tt.method, tt.path); got != tt.public {
				t.Fatalf("isPublicCatalogRead() = %t, want %t", got, tt.public)
			}
		})
	}
}
