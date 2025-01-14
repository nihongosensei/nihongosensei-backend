package elasticsearch

import (
	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchDictionary struct {
	TypedClient *elasticsearch8.TypedClient
}
