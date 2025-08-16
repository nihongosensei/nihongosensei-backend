package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v9"
)

type ElasticsearchDictionary struct {
	TypedClient *elasticsearch.TypedClient
}
