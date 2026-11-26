package resources

import (
	"net/http"
	"strings"

	"github.com/supertokens/supertokens-golang/recipe/session"
)

func SessionMiddleware(next http.Handler, authPath string) http.Handler {
	verified := session.VerifySession(nil, next.ServeHTTP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSuffix(r.URL.Path, "/")
		if path == "/docs" || path == "/openapi.json" ||
			path == authPath || strings.HasPrefix(path, authPath+"/") ||
			isPublicWebhook(r.Method, path) ||
			isPublicCatalogRead(r.Method, path) ||
			!strings.HasPrefix(path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		verified(w, r)
	})
}

func isPublicWebhook(method, path string) bool {
	return method == http.MethodPost && path == "/api/square/webhook"
}

func isPublicCatalogRead(method, path string) bool {
	if method != http.MethodGet {
		return false
	}
	for _, resourcePath := range []string{
		"/api/products",
		"/api/product-variations",
		"/api/product-options",
		"/api/option-types",
		"/api/option-values",
		"/api/order-statuses",
	} {
		if path == resourcePath || strings.HasPrefix(path, resourcePath+"/") {
			return true
		}
	}
	return false
}
