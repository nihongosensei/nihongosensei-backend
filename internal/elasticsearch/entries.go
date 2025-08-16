package elasticsearch

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
)

func (d *ElasticsearchDictionary) SelectEntries(filter dictionary.EntryFilter) (*dictionary.SelectEntries, error) {
	query := &types.Query{
		Bool: &types.BoolQuery{},
	}

	if len(filter.Query) > 0 {
		lowercaseQuery := strings.ToLower(filter.Query)
		// rank exact matches higher
		exactMatchBoost := float32(10)

		query.Bool.Must = append(query.Bool.Must, types.Query{
			Bool: &types.BoolQuery{
				Should: []types.Query{
					{
						Match: map[string]types.MatchQuery{
							"text_japanese": {
								Query: filter.Query,
								Boost: &exactMatchBoost,
							},
						},
					},
					{
						Match: map[string]types.MatchQuery{
							"text_japanese": {
								Query: filter.Query,
								Boost: &exactMatchBoost,
							},
						},
					},
					{
						Term: map[string]types.TermQuery{
							"text_german.keyword": {
								Value: lowercaseQuery,
							},
						},
					},
					{
						Term: map[string]types.TermQuery{
							"text_japanese.keyword": {
								Value: lowercaseQuery,
							},
						},
					},
				},
				MinimumShouldMatch: 1,
			},
		})
		// rank matches with jlpt level slightly higher
		// A value of 10 might prefer jlpt matches over exact matches. 5 seems to be
		// reasonable for now.
		jlptBoost := float32(5)
		query.Bool.Should = append(query.Bool.Should, types.Query{
			Term: map[string]types.TermQuery{
				"has_jlpt_level": {
					Value: true,
					Boost: &jlptBoost,
				},
			},
		})

	}
	if filter.Jlpt != 0 {
		query.Bool.Filter = append(query.Bool.Filter, types.Query{
			Term: map[string]types.TermQuery{
				"jlpt": {
					Value: filter.Jlpt,
				},
			},
		})
	}
	if len(filter.Ids) > 0 {
		query.Bool.Filter = append(query.Bool.Filter, types.Query{
			Terms: &types.TermsQuery{
				TermsQuery: map[string]types.TermsQueryField{
					"id": filter.Ids,
				},
			},
		})
	}
	if len(filter.Kanjis) > 0 {
		query.Bool.Filter = append(query.Bool.Filter, types.Query{
			Terms: &types.TermsQuery{
				TermsQuery: map[string]types.TermsQueryField{
					"kanjis": filter.Kanjis,
				},
			},
		})
	}
	if len(filter.Text) > 0 {
		query.Bool.Filter = append(query.Bool.Filter, types.Query{
			Bool: &types.BoolQuery{
				Should: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"text_german.keyword": {
								Value: filter.Text,
							},
						},
					},
					{
						Term: map[string]types.TermQuery{
							"text_japanese.keyword": {
								Value: filter.Text,
							},
						},
					},
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
		Index("entries").
		Request(req).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	entries := make([]dictionary.Entry, len(res.Hits.Hits))

	for i, hit := range res.Hits.Hits {
		var entry dictionary.Entry

		err = json.Unmarshal(hit.Source_, &entry)
		if err != nil {
			return nil, err
		}

		entries[i] = entry
	}

	return &dictionary.SelectEntries{
		Entries: entries,
		Count:   int(res.Hits.Total.Value),
	}, nil
}
