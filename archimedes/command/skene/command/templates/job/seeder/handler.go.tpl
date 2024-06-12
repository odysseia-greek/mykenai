package seeder

import (
	"bytes"
	"fmt"
	elastic "github.com/odysseia-greek/agora/aristoteles"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/agora/plato/models"
	ptolemaios "github.com/odysseia-greek/delphi/ptolemaios/diplomat"
	"strings"
	"sync"
)

type {{.CapitalizedName}}Handler struct {
	Index      string
	Elastic    elastic.Client
	PolicyName string
	Created    int
	Ambassador *ptolemaios.ClientAmbassador
}

func (s *{{.CapitalizedName}}Handler) DeleteIndexAtStartUp() error {
	deleted, err := s.Elastic.Index().Delete(s.Index)
	logging.Info(fmt.Sprintf("deleted index: %s success: %v", d.Index, deleted))
	if err != nil {
		if deleted {
			return nil
		}
		if strings.Contains(err.Error(), "index_not_found_exception") {
			logging.Error(err.Error())
			return nil
		}

		return err
	}

	return nil
}

func (s *{{.CapitalizedName}}Handler) CreateIndexAtStartup() error {
	logging.Info(fmt.Sprintf("creating policy: %s", s.PolicyName))
	err := s.createPolicyAtStartup()
	if err != nil {
		return err
	}
	logging.Info(fmt.Sprintf("creating index: %s", s.Index))
	query := {{.Index}}Index(s.PolicyName)
	res, err := d.Elastic.Index().Create(s.Index, query)
	if err != nil {
		return err
	}

	logging.Info(fmt.Sprintf("created index: %s", res.Index))
	return nil
}

func (s *{{.CapitalizedName}}Handler) createPolicyAtStartup() error {
	policyCreated, err := s.Elastic.Policy().CreateHotPolicy(s.PolicyName)
	if err != nil {
		return err
	}

	logging.Info(fmt.Sprintf("created policy: %s %v", s.PolicyName, policyCreated.Acknowledged))

	return nil
}

func (s *{{.CapitalizedName}}Handler) AddDirectoryToElastic(example models.ExampleModel, wg *sync.WaitGroup) {
	defer wg.Done()
	var buf bytes.Buffer

	var currBatch int

	for _, ex := range example.Examples {
		currBatch++

		meta := []byte(fmt.Sprintf(`{ "index": {} }%s`, "\n"))
		jsonifiedExample, _ := ex.Marshal()
		jsonifiedExample = append(jsonifiedExample, "\n"...)
		buf.Grow(len(meta) + len(jsonifiedExample))
		buf.Write(meta)
		buf.Write(jsonifiedExample)

		if currBatch == len(example.Examples) {
			res, err := s.Elastic.Document().Bulk(buf, s.Index)
			if err != nil {
				logging.Error(err.Error())
				return
			}

			s.Created = s.Created + len(res.Items)
		}
	}
}
