package main

import (
	"context"
	"fmt"
	"github.com/alexedwards/scs/mongodbstore"
	"github.com/alexedwards/scs/v2"
	"github.com/normegil/evevulcan/internal/config"
	"github.com/normegil/evevulcan/internal/db"
	"github.com/normegil/evevulcan/internal/eveapi"
	"github.com/normegil/evevulcan/internal/http"
	"github.com/normegil/evevulcan/internal/http/middleware"
	"github.com/normegil/evevulcan/ui/web"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	stdhttp "net/http"
)

func main() {
	webFrontend, err := web.Frontend()
	if err != nil {
		panic(fmt.Errorf("load frontend: %w", err))
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongoDBURL()))
	if err != nil {
		panic(fmt.Errorf("connect to mongo using '%s': %w", config.MongoDBURL(), err))
	}
	defer func() {
		if err := client.Disconnect(context.Background()); nil != err {
			log.Error().Err(err).Msg("Could not close resource")
		}
	}()
	mongoDatabase := client.Database("eve-vulcan")

	dbInstance := db.New(mongoDatabase)
	if err = dbInstance.CreateIndexes(); nil != err {
		panic(err)
	}

	api := eveapi.SSO{
		DomainName:  config.EveSSODomainName(),
		Client:      config.EveSSOClientAuth(),
		RedirectURL: *config.EveSSORedirectURL(),
		DB:          dbInstance,
	}

	sessionManager := scs.New()
	sessionManager.Store = mongodbstore.New(mongoDatabase)
	frontendBaseURL := *config.FrontendBaseURL()
	routes, err := http.Routes(frontendBaseURL, webFrontend, dbInstance, api, sessionManager)
	if err != nil {
		panic(fmt.Errorf("load routes: %w", err))
	}

	routes = middleware.SessionIdentityHandler{
		SessionManager: sessionManager,
		DB:             dbInstance,
		ErrHandler:     http.ErrorHandler{},
		Handler:        routes,
	}
	routes = sessionManager.LoadAndSave(routes)
	routes = middleware.AnonymousUserSetter{Handler: routes}
	if config.EnableCrossOriginHeader() {
		routes = middleware.CrossOriginRessourceSharing{Handler: routes, Frontend: frontendBaseURL}
	}
	routes = middleware.RequestLogger{Handler: routes}

	server := stdhttp.Server{
		Addr:    ":18080",
		Handler: routes,
	}

	log.Info().Str("address", server.Addr).Msg("server listening")
	if err := server.ListenAndServe(); nil != err {
		panic(err)
	}
}
