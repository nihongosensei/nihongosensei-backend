package elasticsearch

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/sortorder"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
)

func (d *ElasticsearchDictionary) SelectEntries(ctx context.Context, filter dictionary.EntryFilter) (*dictionary.SelectEntries, error) {
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
							"text_german": {
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
		Sort:  entrySort(filter),
	}

	reqJson, _ := json.Marshal(req)
	d.Logger.Debug("performing elasticsearch search request", "request", reqJson)

	res, err := d.TypedClient.Search().
		Index("entries").
		Request(req).
		Do(ctx)

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

// entrySort translates the requested sorting into elasticsearch sort options.
// Without any criteria the default relevance ordering is kept. Otherwise, ties
// are broken by relevance (if searching) and finally by id, so that paging
// through results is stable.
func entrySort(filter dictionary.EntryFilter) []types.SortCombinations {
	if len(filter.Sort) == 0 {
		return nil
	}

	var sort []types.SortCombinations
	hasScore, hasId := false, false

	for _, s := range filter.Sort {
		switch s.Field {
		case dictionary.EntrySortScore:
			hasScore = true
			order := sortorder.Desc
			if s.Order == dictionary.SortAsc {
				order = sortorder.Asc
			}
			sort = append(sort, types.SortOptions{Score_: &types.ScoreSort{Order: &order}})
		default:
			hasId = hasId || s.Field == dictionary.EntrySortId
			order := sortorder.Asc
			if s.Order == dictionary.SortDesc {
				order = sortorder.Desc
			}
			sort = append(sort, types.SortOptions{
				SortOptions: map[string]types.FieldSort{string(s.Field): {Order: &order}},
			})
		}
	}

	if len(filter.Query) > 0 && !hasScore {
		desc := sortorder.Desc
		sort = append(sort, types.SortOptions{Score_: &types.ScoreSort{Order: &desc}})
	}
	if !hasId {
		asc := sortorder.Asc
		sort = append(sort, types.SortOptions{
			SortOptions: map[string]types.FieldSort{"id": {Order: &asc}},
		})
	}

	return sort
}
