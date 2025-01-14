package elasticsearch

import (
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
	"log/slog"
)

func (d *ElasticsearchDictionary) SelectKanji(kanjiChar string) (*dictionary.Kanji, error) {
	res, err := d.TypedClient.Core.Get("kanjis", kanjiChar).Do(context.Background())

	if err != nil {
		return nil, err
	}
	if !res.Found {
		return nil, dictionary.ErrEntryNotFound
	}

	var kanji dictionary.Kanji
	err = json.Unmarshal(res.Source_, &kanji)

	if err != nil {
		return nil, err
	}

	return &kanji, nil
}

func (d *ElasticsearchDictionary) SelectKanjis(kanjis []string) (*[]dictionary.Kanji, error) {
	limit := len(kanjis)

	req := &search.Request{
		Query: &types.Query{
			Terms: &types.TermsQuery{
				TermsQuery: map[string]types.TermsQueryField{
					"kanji": kanjis,
				},
			},
		},
		Size: &limit,
	}

	reqJson, _ := json.Marshal(req)
	slog.Debug("performing elasticsearch search request", "request", reqJson)

	res, err := d.TypedClient.Search().
		Index("kanjis").
		Request(req).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	k := make([]dictionary.Kanji, len(res.Hits.Hits))

	for i, hit := range res.Hits.Hits {
		var kanji dictionary.Kanji

		err = json.Unmarshal(hit.Source_, &kanji)
		if err != nil {
			return nil, err
		}

		k[i] = kanji
	}

	return &k, nil
}
