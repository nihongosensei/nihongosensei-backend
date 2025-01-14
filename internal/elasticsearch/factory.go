package elasticsearch

import (
	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/nihongosensei/nihongosensei-backend/internal/environment"
)

func CreateClientFromEnv() (*elasticsearch8.TypedClient, error) {
	host := environment.Require("ELASTICSEARCH_HOST")
	username := environment.Require("ELASTICSEARCH_USERNAME")
	password := environment.Require("ELASTICSEARCH_PASSWORD")

	esClient, err := elasticsearch8.NewTypedClient(elasticsearch8.Config{
		Addresses: []string{host},
		Username:  username,
		Password:  password,
	})

	return esClient, err
}
