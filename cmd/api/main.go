package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/neh333/KaininAPIV2/config"
	"github.com/neh333/KaininAPIV2/internal/transport/http/resources"
	squarehandler "github.com/neh333/KaininAPIV2/internal/transport/http/square"
	"github.com/sirupsen/logrus"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/supertokens"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

func main() {
	serverConfig, err := config.LoadConfig()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load config")
	}
	if err := initializeSuperTokens(serverConfig); err != nil {
		logrus.WithError(err).Fatal("Failed to initialize SuperTokens")
	}

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create the Azure credential")
	}

	client, err := azsecrets.NewClient(
		"https://kaininkv.vault.azure.net/",
		credential,
		nil,
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create the Key Vault client")
	}
	connectionStringResponse, err := client.GetSecret(
		context.Background(),
		"KaininStoreUSDBConnString",
		"",
		nil,
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get the database connection string")
	}
	if connectionStringResponse.Value == nil {
		logrus.Fatal("The database connection string secret is empty")
	}

	db, err := gorm.Open(sqlserver.Open(*connectionStringResponse.Value), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to SQL Server")
	}
	logrus.Info("Successfully connected to SQL Server")

	router := chi.NewRouter()
	router.Use(middleware.Logger, middleware.Recoverer)
	apiVersion := serverConfig.Version
	if apiVersion == "" {
		apiVersion = "1.0.0"
	}
	apiConfig := huma.DefaultConfig("Kainin API", apiVersion)
	apiConfig.DocsRenderer = huma.DocsRendererScalar
	api := humachi.New(router, apiConfig)
	resources.RegisterRoutes(api, db)
	squarehandler.RegisterRoutes(
		router,
		db,
		serverConfig.SquareWebhookSignatureKey,
		serverConfig.SquareWebhookNotificationURL,
	)

	port := serverConfig.Port
	if port == "" {
		port = "8888"
	}
	authBasePath := serverConfig.SupertokensAPIBasePath
	if authBasePath == "" {
		authBasePath = "/auth"
	}
	server := &http.Server{
		Addr: ":" + port,
		Handler: corsMiddleware(
			supertokens.Middleware(resources.SessionMiddleware(router, authBasePath)),
			serverConfig.AllowedOrigins,
		),
		ReadHeaderTimeout: 10 * time.Second,
	}

	logrus.WithFields(logrus.Fields{
		"env":     serverConfig.Env,
		"version": apiVersion,
		"port":    port,
	}).Info("Starting API server")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logrus.WithError(err).Fatal("API server stopped unexpectedly")
	}
}

func initializeSuperTokens(cfg config.ConfigSchema) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "SUPERTOKENS_CONNECTION_URI", value: cfg.SupertokensConnectionURI},
		{name: "SUPERTOKENS_APP_NAME", value: cfg.SupertokensAppName},
		{name: "SUPERTOKENS_API_DOMAIN", value: cfg.SupertokensAPIDomain},
		{name: "SUPERTOKENS_WEBSITE_DOMAIN", value: cfg.SupertokensWebsiteDomain},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s must be configured", field.name)
		}
	}

	apiBasePath := cfg.SupertokensAPIBasePath
	if apiBasePath == "" {
		apiBasePath = "/auth"
	}
	websiteBasePath := cfg.SupertokensWebBasePath
	if websiteBasePath == "" {
		websiteBasePath = "/auth"
	}
	return supertokens.Init(supertokens.TypeInput{
		Supertokens: &supertokens.ConnectionInfo{
			ConnectionURI: cfg.SupertokensConnectionURI,
			APIKey:        cfg.SupertokensAPIKey,
		},
		AppInfo: supertokens.AppInfo{
			AppName:         cfg.SupertokensAppName,
			APIDomain:       cfg.SupertokensAPIDomain,
			WebsiteDomain:   cfg.SupertokensWebsiteDomain,
			APIBasePath:     &apiBasePath,
			WebsiteBasePath: &websiteBasePath,
		},
		RecipeList: []supertokens.Recipe{
			emailpassword.Init(nil),
			session.Init(nil),
		},
	})
}

func corsMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && containsOrigin(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(
				append([]string{"Content-Type", "Authorization"}, supertokens.GetAllCORSHeaders()...),
				",",
			))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func containsOrigin(origins []string, origin string) bool {
	for _, allowed := range origins {
		if allowed == origin {
			return true
		}
	}
	return false
}
