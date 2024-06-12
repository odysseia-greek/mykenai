package seeder

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/odysseia-greek/agora/archytas"
	"github.com/odysseia-greek/agora/aristoteles"
	"github.com/odysseia-greek/agora/aristoteles/models"
	"github.com/odysseia-greek/agora/plato/config"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/agora/plato/service"
	aristophanes "github.com/odysseia-greek/attike/aristophanes/comedy"
	pbar "github.com/odysseia-greek/attike/aristophanes/proto"
	"github.com/odysseia-greek/delphi/ptolemaios/diplomat"
	pb "github.com/odysseia-greek/delphi/ptolemaios/proto"
	aristarchos "github.com/odysseia-greek/olympia/aristarchos/scholar"
	"google.golang.org/grpc/metadata"
	"os"
	"time"
)

const (
	defaultIndex string = "{{.Index}}"
)

func CreateNewConfig() (*{{.CapitalizedName}}Handler, error) {
    tls := config.BoolFromEnv(config.EnvTlSKey)

    var cfg models.Config
    ambassador := diplomat.NewClientAmbassador()

    healthy := ambassador.WaitForHealthyState()
    if !healthy {
        logging.Info("tracing service not ready - restarting seems the only option")
        os.Exit(1)
    }

    traceId := uuid.New().String()
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
    defer cancel()
    md := metadata.New(map[string]string{service.HeaderKey: traceId})
    ctx = metadata.NewOutgoingContext(context.Background(), md)
    vaultConfig, err := ambassador.GetSecret(ctx, &pbp.VaultRequest{})
    if err != nil {
        logging.Error(err.Error())
        return nil, nil, err
    }

    elasticService := aristoteles.ElasticService(tls)

    cfg = models.Config{
        Service:     elasticService,
        Username:    vaultConfig.ElasticUsername,
        Password:    vaultConfig.ElasticPassword,
        ElasticCERT: vaultConfig.ElasticCERT,
    }

    elastic, err := aristoteles.NewClient(cfg)
    if err != nil {
        return nil, nil, err
    }

	index := config.StringFromEnv(config.EnvIndex, defaultIndex)
    policyName := fmt.Sprintf("%s_policy", index)

	return &{{.CapitalizedName}}Handler{
		Index:      index,
        PolicyName: policyName,
		Elastic:    elastic,
		Ambassador: ambassador,
		Created: 0,
	}, nil
}
