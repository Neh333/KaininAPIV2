package resources

import (
	"net/http"
	"testing"
)

func TestSquareWebhookIsPublicOnlyForPost(t *testing.T) {
	if !isPublicWebhook(http.MethodPost, "/api/square/webhook") {
		t.Fatal("expected POST Square webhook to bypass session authentication")
	}
	if isPublicWebhook(http.MethodGet, "/api/square/webhook") {
		t.Fatal("expected non-POST Square webhook requests to remain protected")
	}
}
