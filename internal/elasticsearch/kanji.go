package elasticsearch

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
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

func (d *ElasticsearchDictionary) SelectKanjis(filter dictionary.KanjiFilters) (*dictionary.SelecKanjis, error) {
	query := &types.Query{
		Bool: &types.BoolQuery{},
	}

	if len(filter.Kanjis) > 0 {
		query.Bool.Must = append(query.Bool.Must, types.Query{
			Terms: &types.TermsQuery{
				TermsQuery: map[string]types.TermsQueryField{
					"kanji": filter.Kanjis,
				},
			},
		})
	}
	if len(filter.Strokes) > 0 {
		query.Bool.Must = append(query.Bool.Must, types.Query{
			Terms: &types.TermsQuery{
				TermsQuery: map[string]types.TermsQueryField{
					"strokes": filter.Strokes,
				},
			},
		})
	}
	if len(filter.Grade) > 0 {
		query.Bool.Must = append(query.Bool.Must, types.Query{
			Terms: &types.TermsQuery{
				TermsQuery: map[string]types.TermsQueryField{
					"grade": filter.Grade,
				},
			},
		})
	}
	if len(filter.Radical) > 0 {
		query.Bool.Must = append(query.Bool.Must, types.Query{
			Terms: &types.TermsQuery{
				TermsQuery: map[string]types.TermsQueryField{
					"radical": filter.Radical,
				},
			},
		})
	}

	limit := filter.Limit
	offset := filter.Offset

	req := &search.Request{
		Query: query,
		Size:  &limit,
		From:  &offset,
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

	return &dictionary.SelecKanjis{
		Kanjis: k,
		Count:  int(res.Hits.Total.Value),
	}, nil
}
