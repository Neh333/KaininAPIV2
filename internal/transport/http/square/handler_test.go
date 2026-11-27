package square

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/neh333/KaininAPIV2/internal/service"
	squareSDK "github.com/square/square-go-sdk/v5"
	squarecore "github.com/square/square-go-sdk/v5/core"
	squarewebhookclient "github.com/square/square-go-sdk/v5/webhooks/client"
)

const (
	testWebhookBody = `{"merchant_id":"MLEFBHHSJGVHD","type":"webhooks.test_notification","event_id":"ac3ac95b-f97d-458c-a6e6-18981597e05f","created_at":"2022-07-13T20:30:59.037339943Z","data":{"type":"webhooks","id":"bc368e64-01aa-407e-b46e-3231809b1129"}}`
	testWebhookURL  = "https://webhook.site/679a4f3a-dcfa-49ee-bac5-9d0edad886b9"
	testWebhookKey  = "Ibxx_5AKakO-3qeNVR61Dw"
	testSignature   = "GF4YkrJgGBDZ9NIYbNXBnMzqb2HoL4RW/S6vkZ9/2N4="
)

func TestReceiveWebhookVerifiesSignatureWithSquareSDK(t *testing.T) {
	handler := testHandler(testWebhookKey, testWebhookURL)
	request := httptest.NewRequest(http.MethodPost, webhookPath, strings.NewReader(testWebhookBody))
	request.Header.Set("x-square-hmacsha256-signature", testSignature)
	response := httptest.NewRecorder()

	handler.ReceiveWebhook(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestReceiveWebhookRejectsInvalidSignature(t *testing.T) {
	handler := testHandler(testWebhookKey, testWebhookURL)
	request := httptest.NewRequest(http.MethodPost, webhookPath, strings.NewReader(testWebhookBody))
	request.Header.Set("x-square-hmacsha256-signature", "invalid")
	response := httptest.NewRecorder()

	handler.ReceiveWebhook(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestReceiveWebhookRejectsNotificationURLMismatch(t *testing.T) {
	handler := testHandler(testWebhookKey, testWebhookURL+"/wrong")
	request := httptest.NewRequest(http.MethodPost, webhookPath, strings.NewReader(testWebhookBody))
	request.Header.Set("x-square-hmacsha256-signature", testSignature)
	response := httptest.NewRecorder()

	handler.ReceiveWebhook(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestReceiveWebhookFailsClosedWithoutConfiguration(t *testing.T) {
	handler := testHandler("", "")
	request := httptest.NewRequest(http.MethodPost, webhookPath, strings.NewReader(testWebhookBody))
	response := httptest.NewRecorder()

	handler.ReceiveWebhook(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}
}

func TestValidateShippingAddressNormalizesInput(t *testing.T) {
	address := shippingAddress{
		AddressLine1: "  123 Main St  ",
		Locality:     " Seattle ",
		PostalCode:   " 98101 ",
		Country:      " us ",
	}

	if err := validateShippingAddress(&address); err != nil {
		t.Fatalf("expected valid address, got error: %v", err)
	}
	if address.AddressLine1 != "123 Main St" || address.Locality != "Seattle" ||
		address.PostalCode != "98101" || address.Country != "US" {
		t.Fatalf("address was not normalized: %+v", address)
	}
}

func TestValidateShippingAddressRejectsMissingFieldsAndInvalidCountry(t *testing.T) {
	tests := []struct {
		name    string
		address shippingAddress
	}{
		{
			name: "missing street",
			address: shippingAddress{
				Locality: "Seattle", PostalCode: "98101", Country: "US",
			},
		},
		{
			name: "invalid country",
			address: shippingAddress{
				AddressLine1: "123 Main St", Locality: "Seattle",
				PostalCode: "98101", Country: "USA",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateShippingAddress(&tt.address); err == nil {
				t.Fatal("expected invalid address to be rejected")
			}
		})
	}
}

func TestAddressFromSquareHandlesOptionalFields(t *testing.T) {
	address := addressFromSquare(&squareSDK.Address{})
	if address == nil {
		t.Fatal("expected an address response")
	}
	if address.AddressLine1 != "" || address.Country != "" {
		t.Fatalf("expected empty optional fields, got %+v", address)
	}
}

func testHandler(signatureKey, notificationURL string) *Handler {
	return &Handler{
		service:         service.NewSquareService(nil),
		signatureKey:    signatureKey,
		notificationURL: notificationURL,
		verifier:        squarewebhookclient.NewClient(squarecore.NewRequestOptions()),
	}
}
