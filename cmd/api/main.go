package main

import (
	"context"
	"go-api-template/config"

	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"

	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/supertokens"

	"github.com/sirupsen/logrus"
)

func main() {
	apiBasePath := "/auth"
	websiteBasePath := "/auth"

	err := supertokens.Init(supertokens.TypeInput{
		Supertokens: &supertokens.ConnectionInfo{
			// We use try.supertokens for demo purposes.
			// At the end of the tutorial we will show you how to create
			// your own SuperTokens core instance and then update your config.
			ConnectionURI: "https://try.supertokens.io",
			// APIKey: <YOUR_API_KEY>
		},
		AppInfo: supertokens.AppInfo{
			AppName:         "<YOUR_APP_NAME>",
			APIDomain:       "<YOUR_API_DOMAIN>",
			WebsiteDomain:   "<YOUR_WEBSITE_DOMAIN>",
			APIBasePath:     &apiBasePath,
			WebsiteBasePath: &websiteBasePath,
		},
		RecipeList: []supertokens.Recipe{
			emailpassword.Init(nil),
			session.Init(nil),
		},
	})

	http.ListenAndServe("SERVER ADDRESS", corsMiddleware(
		supertokens.Middleware(http.HandlerFunc(func(rw http.ResponseWriter,
			r *http.Request) {
			// TODO: Handle your APIs..

		}))))

	if err != nil {
		panic(err.Error())
	}

	serverConfig, err := config.LoadConfig()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load config")
		panic(err)
	}

	client, err := azsecrets.NewClient("https://kaininkv.vault.azure.net/", &azidentity.DefaultAzureCredential{}, nil)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create the client: %v")
	}

	connectionString, err := client.GetSecret(context.TODO(), "KaininStoreUSDBConnString", "", nil)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get secret: %v")
		panic(err)
	}

	db, err := sqlx.Connect("sqlserver", *connectionString.Value)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
		panic(err)
	}
	defer db.Close()

	logrus.WithFields(logrus.Fields{
		"env":     serverConfig.Env,
		"version": serverConfig.Version,
		"port":    serverConfig.Port,
	}).Info("Loaded app config")

	// ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer stop()

	// server := internal.NewServer()
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		server.OnShutdown()
	// 	}
	// }()

	// if err := server.Run(ctx); err != nil && ctx.Err() == nil {
	// 	logrus.WithError(err).Fatal("Server exited with error")
	// }
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, r *http.Request) {
		response.Header().Set("Access-Control-Allow-Origin", "<YOUR_WEBSITE_DOMAIN>")
		response.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			// we add content-type + other headers used by SuperTokens
			response.Header().Set("Access-Control-Allow-Headers",
				strings.Join(append([]string{"Content-Type"},
					supertokens.GetAllCORSHeaders()...), ","))
			response.Header().Set("Access-Control-Allow-Methods", "*")
			response.Write([]byte(""))
		} else {
			next.ServeHTTP(response, r)
		}
	})
}
