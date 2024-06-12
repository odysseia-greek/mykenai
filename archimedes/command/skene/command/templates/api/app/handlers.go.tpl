package app

import (
	"github.com/odysseia-greek/agora/plato/middleware"
	"github.com/odysseia-greek/agora/plato/models"
	"net/http"
	"time"
)

type {{.CapitalizedName}}Handler struct {
	Elastic    aristoteles.Client
	Index      string
	Streamer   pb.TraceService_ChorusClient
	Cancel     context.CancelFunc
}

// PingPong pongs the ping
func (a *{{.CapitalizedName}}Handler) pingPong(w http.ResponseWriter, req *http.Request) {
	pingPong := models.ResultModel{Result: "pong"}
	middleware.ResponseWithJson(w, pingPong)
}

// returns the health of the api
func (a *{{.CapitalizedName}}Handler) health(w http.ResponseWriter, req *http.Request) {
	elasticHealth := a.Elastic.Health().Info()
	dbHealth := models.DatabaseHealth{
		Healthy:       elasticHealth.Healthy,
		ClusterName:   elasticHealth.ClusterName,
		ServerName:    elasticHealth.ServerName,
		ServerVersion: elasticHealth.ServerVersion,
	}

	healthy := models.Health{
		Healthy:  dbHealth.Healthy,
		Time:     time.Now().String(),
		Database: dbHealth,
	}
	if !healthy.Healthy {
		middleware.ResponseWithCustomCode(w, http.StatusBadGateway, healthy)
		return
	}

	middleware.ResponseWithJson(w, healthy)
}

// Example
func (a *{{.CapitalizedName}}Handler) exampleEndpoint(w http.ResponseWriter, req *http.Request) {
    var requestId string
	fromContext := req.Context().Value(config.DefaultTracingName)
	if fromContext == nil {
		requestId = req.Header.Get(config.HeaderKey)
	} else {
		requestId = fromContext.(string)
	}
	splitID := strings.Split(requestId, "+")

	traceCall := false
	var traceID, spanID string

	if len(splitID) >= 3 {
		traceCall = splitID[2] == "1"
	}

	if len(splitID) >= 1 {
		traceID = splitID[0]
	}
	if len(splitID) >= 2 {
		spanID = splitID[1]
	}

    query := textAggregationQuery()

    elasticResult, err := h.Elastic.Query().MatchRaw(h.Index, query)
    if err != nil {
        e := models.ElasticSearchError{
            ErrorModel: models.ErrorModel{UniqueCode: requestId},
            Message: models.ElasticErrorMessage{
                ElasticError: err.Error(),
            },
        }
        middleware.ResponseWithJson(w, e)
        return
    }

    var agg map[string]interface{}
    err = json.Unmarshal(elasticResult, &agg)
    if err != nil {
        e := models.ValidationError{
            ErrorModel: models.ErrorModel{UniqueCode: requestId},
            Messages: []models.ValidationMessages{
                {
                    Field:   "unmarshall action failed internally",
                    Message: err.Error(),
                },
            },
        }
        middleware.ResponseWithJson(w, e)
        return
    }

    middleware.ResponseWithCustomCode(w, http.StatusOK, agg)
}

func textAggregationQuery() map[string]interface{} {
	return map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"authors": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "author.keyword",
					"size":  100,
				},
				"aggs": map[string]interface{}{
					"books": map[string]interface{}{
						"terms": map[string]interface{}{
							"field": "book.keyword",
							"size":  100,
						},
						"aggs": map[string]interface{}{
							"references": map[string]interface{}{
								"terms": map[string]interface{}{
									"field": "reference",
									"size":  100,
								},
								"aggs": map[string]interface{}{
									"sections": map[string]interface{}{
										"nested": map[string]interface{}{
											"path": "rhemai",
										},
										"aggs": map[string]interface{}{
											"section_ids": map[string]interface{}{
												"terms": map[string]interface{}{
													"field": "rhemai.section",
													"size":  100,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}