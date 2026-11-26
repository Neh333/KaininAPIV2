package square

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/neh333/KaininAPIV2/internal/service"
	"github.com/sirupsen/logrus"
	squareSDK "github.com/square/square-go-sdk/v5"
	squarecore "github.com/square/square-go-sdk/v5/core"
	squarewebhookclient "github.com/square/square-go-sdk/v5/webhooks/client"
	"gorm.io/gorm"
)

const (
	webhookPath    = "/api/square/webhook"
	maxWebhookBody = 2 << 20
)

type Handler struct {
	service         *service.SquareService
	signatureKey    string
	notificationURL string
	verifier        *squarewebhookclient.Client
}

func RegisterRoutes(router chi.Router, db *gorm.DB, signatureKey, notificationURL string) {
	handler := &Handler{
		service:         service.NewSquareService(db),
		signatureKey:    strings.TrimSpace(signatureKey),
		notificationURL: strings.TrimSpace(notificationURL),
		verifier:        squarewebhookclient.NewClient(squarecore.NewRequestOptions()),
	}
	router.Post(webhookPath, handler.ReceiveWebhook)
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
