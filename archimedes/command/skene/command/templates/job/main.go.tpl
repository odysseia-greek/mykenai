package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/odysseia-greek/agora/plato/logging"
	pb "github.com/odysseia-greek/delphi/ptolemaios/proto"
	"github.com/odysseia-greek/{{.RepoName}}/{{.Name}}/seeder"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

var documents int

//go:embed {{.Embed}}
var {{.Embed}} embed.FS

func main() {
	//https://patorjk.com/software/taag/#p=display&f=Crawford2&t={{.Name}}
	logging.System(`

`)
	logging.System("\"INSERT GREEK QUOTE\"")
	logging.System("\"INSERT TRANSLATION\"")
	logging.System("starting up.....")
	logging.System("starting up and getting env variables")

	handler, err := seeder.CreateNewConfig()
	if err != nil {
		logging.Error(err.Error())
		log.Fatal("death has found me")
	}

	root := "{{.Embed}}"
	rootDir, err := {{.Embed}}.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}

	err = handler.DeleteIndexAtStartUp()
	if err != nil {
		log.Fatal(err)
	}
	err = handler.CreateIndexAtStartup()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	for _, dir := range rootDir {
		logging.Debug("working on the following directory: " + dir.Name())
		if dir.IsDir() {
			filePath := path.Join(root, dir.Name())
			files, err := {{.Embed}}.ReadDir(filePath)
			if err != nil {
				log.Fatal(err)
			}
			for _, f := range files {
				logging.Debug(fmt.Sprintf("found %s in %s", f.Name(), filePath))
				plan, _ := {{.Embed}}.ReadFile(path.Join(filePath, f.Name()))
				var someModel seeder.ExampleModel
				err := json.Unmarshal(plan, &someModel)
				if err != nil {
					log.Fatal(err)
				}

				documents += len(someModel)

				wg.Add(1)
				go handler.AddDirectoryToElastic(someModel, &wg)
			}
		}
	}

	wg.Wait()

	logging.Info(fmt.Sprintf("created: %s", strconv.Itoa(handler.Created)))
	logging.Info(fmt.Sprintf("texts found in rhema: %s", strconv.Itoa(documents)))

	logging.Debug("closing ptolemaios because job is done")
	// just setting a code that could be used later to check is if it was sent from an actual service
	uuidCode := uuid.New().String()
	_, err = handler.Ambassador.ShutDown(context.Background(), &pb.ShutDownRequest{Code: uuidCode})
	if err != nil {
		logging.Error(err.Error())
	}
	os.Exit(0)
}
