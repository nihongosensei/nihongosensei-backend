package elasticsearch

import (
	"log/slog"

	"github.com/elastic/go-elasticsearch/v9"
)

type ElasticsearchDictionary struct {
	TypedClient *elasticsearch.TypedClient
	Logger      *slog.Logger
}
