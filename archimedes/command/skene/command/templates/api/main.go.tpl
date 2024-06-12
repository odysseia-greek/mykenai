package main

import (
	"context"
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/{{.RepoName}}/{{.Name}}/app"
	"log"
	"net/http"
	"os"
)

const standardPort = ":5000"

func main() {
	port := os.Getenv("{{.Port}}")
	if port == "" {
		port = standardPort
	}

	//https://patorjk.com/software/taag/#p=display&f=Crawford2&t={{.Name}}
	logging.System(`

`)
	logging.System("\"INSERT GREEK QUOTE\"")
	logging.System("\"INSERT TRANSLATION\"")
	logging.System("starting up.....")
	logging.System("starting up and getting env variables")

	ctx := context.Background()
	apiConfig, err := app.CreateNewConfig(ctx)
	if err != nil {
		logging.Error(err.Error())
		log.Fatal("death has found me")
	}

	srv := text.InitRoutes(apiConfig)

	logging.Info(fmt.Sprintf("%s : %s", "running on port", port))
	err = http.ListenAndServe(port, srv)
	if err != nil {
		panic(err)
	}
}
