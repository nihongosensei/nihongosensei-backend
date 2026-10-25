package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/nihongosensei/nihongosensei-backend/internal/environment"
)

func CreateClientFromEnv() (*elasticsearch.TypedClient, error) {
	host, err := environment.Require("ELASTICSEARCH_HOST")
	if err != nil {
		return nil, err
	}
	username, err := environment.Require("ELASTICSEARCH_USERNAME")
	if err != nil {
		return nil, err
	}
	password, err := environment.Require("ELASTICSEARCH_PASSWORD")
	if err != nil {
		return nil, err
	}

	esClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{host},
		Username:  username,
		Password:  password,
	})

	return esClient, err
}
