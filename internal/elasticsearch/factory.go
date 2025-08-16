package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/nihongosensei/nihongosensei-backend/internal/environment"
)

func CreateClientFromEnv() (*elasticsearch.TypedClient, error) {
	host := environment.Require("ELASTICSEARCH_HOST")
	username := environment.Require("ELASTICSEARCH_USERNAME")
	password := environment.Require("ELASTICSEARCH_PASSWORD")

	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{host},
		Username:  username,
		Password:  password,
	})

	return esClient, err
}
