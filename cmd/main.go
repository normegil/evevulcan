package main

import (
	"context"
	"fmt"
	"github.com/normegil/evevulcan/internal/config"
	"github.com/normegil/evevulcan/internal/http"
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
	database := client.Database("eve-vulcan")

	routes, err := http.Routes(webFrontend, database)
	if err != nil {
		panic(fmt.Errorf("load routes: %w", err))
	}
	server := stdhttp.Server{
		Addr:    ":18080",
		Handler: routes,
	}

	log.Info().Str("address", server.Addr).Msg("server listening")
	if err := server.ListenAndServe(); nil != err {
		panic(err)
	}
}