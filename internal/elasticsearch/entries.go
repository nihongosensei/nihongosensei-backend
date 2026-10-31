package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

	// The field filters are applied as post filter, so that the aggregations
	// can decide on their own which of them to respect.
	fieldFilters := entryFieldFilters(filter)
	if len(fieldFilters) > 0 {
		req.PostFilter = &types.Query{Bool: &types.BoolQuery{Filter: queriesOf(fieldFilters, "")}}
	}
	req.Aggregations = entryAggregations(filter, fieldFilters)

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

	aggregations, err := parseEntryAggregations(filter, res.Aggregations)
	if err != nil {
		return nil, err
	}

	return &dictionary.SelectEntries{
		Entries:      entries,
		Count:        int(res.Hits.Total.Value),
		Aggregations: aggregations,
	}, nil
}

// aggregationValuesName is the name of the terms aggregation nested in the
// filter aggregation of each requested field.
const aggregationValuesName = "values"

// maxAggregationBuckets is high enough to return every value of the supported
// fields (5 jlpt levels, ~25 grammar types).
const maxAggregationBuckets = 50

// entryFieldFilters returns the terms filters of the multi-valued fields,
// keyed by field. Values within a field are combined with OR.
func entryFieldFilters(filter dictionary.EntryFilter) map[dictionary.EntryAggregationField]types.Query {
	filters := map[dictionary.EntryAggregationField]types.Query{}
	if len(filter.Jlpt) > 0 {
		values := make([]types.FieldValue, len(filter.Jlpt))
		for i, v := range filter.Jlpt {
			values[i] = v
		}
		filters[dictionary.EntryAggregationJlpt] = termsQuery("jlpt", values)
	}
	if len(filter.GrammarType) > 0 {
		values := make([]types.FieldValue, len(filter.GrammarType))
		for i, v := range filter.GrammarType {
			values[i] = v
		}
		filters[dictionary.EntryAggregationGrammarType] = termsQuery("grammar_type", values)
	}
	return filters
}

func termsQuery(field string, values []types.FieldValue) types.Query {
	return types.Query{
		Terms: &types.TermsQuery{
			TermsQuery: map[string]types.TermsQueryField{field: values},
		},
	}
}

// queriesOf returns the filters in a stable order, leaving out the given field.
func queriesOf(filters map[dictionary.EntryAggregationField]types.Query, except dictionary.EntryAggregationField) []types.Query {
	var queries []types.Query
	for _, field := range []dictionary.EntryAggregationField{dictionary.EntryAggregationJlpt, dictionary.EntryAggregationGrammarType} {
		if q, ok := filters[field]; ok && field != except {
			queries = append(queries, q)
		}
	}
	return queries
}

// entryAggregations builds one aggregation per requested field. It counts the
// entries matching the query and all field filters except the one on its own
// field, so that unselected values of that field remain visible.
func entryAggregations(filter dictionary.EntryFilter, fieldFilters map[dictionary.EntryAggregationField]types.Query) map[string]types.Aggregations {
	if len(filter.Aggregations) == 0 {
		return nil
	}

	aggs := map[string]types.Aggregations{}
	for _, field := range filter.Aggregations {
		size := maxAggregationBuckets
		name := string(field)
		aggs[name] = types.Aggregations{
			Filter: &types.Query{Bool: &types.BoolQuery{Filter: queriesOf(fieldFilters, field)}},
			Aggregations: map[string]types.Aggregations{
				aggregationValuesName: {
					Terms: &types.TermsAggregation{Field: &name, Size: &size},
				},
			},
		}
	}
	return aggs
}

// parseEntryAggregations converts the terms buckets of the requested
// aggregations. Buckets are ordered by count, as returned by elasticsearch.
func parseEntryAggregations(filter dictionary.EntryFilter, aggs map[string]types.Aggregate) (map[dictionary.EntryAggregationField][]dictionary.AggregationBucket, error) {
	if len(filter.Aggregations) == 0 {
		return nil, nil
	}

	result := map[dictionary.EntryAggregationField][]dictionary.AggregationBucket{}
	for _, field := range filter.Aggregations {
		filtered, ok := aggs[string(field)].(*types.FilterAggregate)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregation type %T for %s", aggs[string(field)], field)
		}

		buckets := []dictionary.AggregationBucket{}
		switch terms := filtered.Aggregations[aggregationValuesName].(type) {
		case *types.LongTermsAggregate:
			list, _ := terms.Buckets.([]types.LongTermsBucket)
			for _, b := range list {
				buckets = append(buckets, dictionary.AggregationBucket{Key: strconv.FormatInt(b.Key, 10), Count: int(b.DocCount)})
			}
		case *types.StringTermsAggregate:
			list, _ := terms.Buckets.([]types.StringTermsBucket)
			for _, b := range list {
				buckets = append(buckets, dictionary.AggregationBucket{Key: fmt.Sprint(b.Key), Count: int(b.DocCount)})
			}
		default:
			return nil, fmt.Errorf("unexpected terms aggregation type %T for %s", terms, field)
		}
		result[field] = buckets
	}
	return result, nil
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
