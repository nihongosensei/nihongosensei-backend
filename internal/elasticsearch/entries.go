package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/operator"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/sortorder"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
	"github.com/nihongosensei/nihongosensei-backend/internal/japanese"
)

func (d *ElasticsearchDictionary) SelectEntries(ctx context.Context, filter dictionary.EntryFilter) (*dictionary.SelectEntries, error) {
	// leading and trailing whitespace would break the exact matches
	filter.Query = strings.TrimSpace(filter.Query)
	filter.Text = strings.TrimSpace(filter.Text)

	query := &types.Query{
		Bool: &types.BoolQuery{},
	}

	if len(filter.Query) > 0 {
		query.Bool.Must = append(query.Bool.Must, entryTextQuery(filter.Query))
		query.Bool.Should = append(query.Bool.Should, frequencyQuery())
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
		// match (instead of term) applies the normalizer of the keyword fields
		query.Bool.Filter = append(query.Bool.Filter, types.Query{
			Bool: &types.BoolQuery{
				Should: []types.Query{
					{Match: map[string]types.MatchQuery{"text_german.keyword": {Query: filter.Text}}},
					{Match: map[string]types.MatchQuery{"text_japanese.keyword": {Query: filter.Text}}},
				},
			},
		})
	}

	limit := filter.Limit
	offset := filter.Offset

	req := &search.Request{
		Query:          query,
		TrackTotalHits: true,
		Size:           &limit,
		From:           &offset,
		Sort:           entrySort(filter),
	}

	// The field filters are applied as post filter, so that the aggregations
	// can decide on their own which of them to respect.
	fieldFilters := entryFieldFilters(filter)
	if len(fieldFilters) > 0 {
		req.PostFilter = &types.Query{Bool: &types.BoolQuery{Filter: queriesOf(fieldFilters, "")}}
	}
	req.Aggregations = entryAggregations(filter, fieldFilters)

	if d.Logger.Enabled(ctx, slog.LevelDebug) {
		reqJson, _ := json.Marshal(req)
		d.Logger.Debug("performing elasticsearch search request", "request", string(reqJson))
	}

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

// Exact hits on the keyword fields have to outrank analyzed matches, which also
// happen to be the only way to find words that the analyzers drop as stop words.
const (
	exactMatchBoost = float32(30)
	textMatchBoost  = float32(2)
	// prefixMatchBoost ranks entries starting with the text above analyzed
	// matches, but below exact ones.
	prefixMatchBoost = float32(5)
	// frequencyBoost is the score of the most frequent words. It decides between
	// equally good matches and can lift a common word above weaker text matches,
	// but not above exact ones.
	frequencyBoost = float32(10)
	// frequencyPivot is the frequency rank which gets half of the frequencyBoost.
	frequencyPivot = float32(5000)
	// unstemmedMatchBoost ranks the exact word form above its stemmed variants.
	unstemmedMatchBoost = float32(1)
)

// frequencyQuery rewards frequent words: the score goes from frequencyBoost
// for the most frequent word down towards 0 and is halved at frequencyPivot.
// Entries without a frequency rank do not get a score from it.
func frequencyQuery() types.Query {
	// the field stores 1/rank, as a lower rank is better, so the pivot has to be
	// the inverse as well
	pivot := float32(1) / frequencyPivot
	return types.Query{
		RankFeature: &types.RankFeatureQuery{
			Field:      "frequency",
			Boost:      new(frequencyBoost),
			Saturation: &types.RankFeatureFunctionSaturation{Pivot: &pivot},
		},
	}
}

// entryTextQuery matches the text against the analyzed text field of its
// language (all terms have to match, so that additional words narrow the
// result) and, with a much higher boost, exactly against the normalized keyword
// fields (the written form, the hiragana reading and the translations).
//
// Text with japanese characters is only searched in the japanese fields, where
// it additionally matches the beginning of the written form and the reading.
// Any other text is searched in the german fields with a constant score (an
// entry has many senses, so counting the occurrences or the length of the field
// would favor words with few senses over common ones), where the unstemmed field
// ranks the exact word form higher. Among equal matches the frequency decides. The exact and prefix matches are scored as
// one constant each, so that an entry matching several keyword fields (e.g.
// orth and text_japanese.keyword hold the same value) is not ranked above
// others. They use match queries, which apply the normalizer of the field to
// the text.
func entryTextQuery(text string) types.Query {
	textMatch := func(field string, boost float32) types.Query {
		return types.Query{Match: map[string]types.MatchQuery{
			field: {Query: text, Boost: new(boost), Operator: &operator.And},
		}}
	}
	constantMatch := func(boost float32, fields ...string) types.Query {
		matches := make([]types.Query, len(fields))
		for i, field := range fields {
			matches[i] = types.Query{Match: map[string]types.MatchQuery{field: {Query: text}}}
		}
		return types.Query{ConstantScore: &types.ConstantScoreQuery{
			Boost:  new(boost),
			Filter: types.Query{Bool: &types.BoolQuery{Should: matches, MinimumShouldMatch: 1}},
		}}
	}

	// constantTextMatch scores a match of all terms the same, no matter how often
	// they occur or how long the field is.
	constantTextMatch := func(field string, boost float32) types.Query {
		return types.Query{ConstantScore: &types.ConstantScoreQuery{
			Boost: new(boost),
			Filter: types.Query{Match: map[string]types.MatchQuery{
				field: {Query: text, Operator: &operator.And},
			}},
		}}
	}

	var should []types.Query
	if japanese.ContainsJapanese(text) {
		should = []types.Query{
			textMatch("text_japanese", textMatchBoost),
			constantMatch(prefixMatchBoost, "reading.prefix", "orth.prefix"),
			constantMatch(exactMatchBoost, "text_japanese.keyword", "orth", "reading"),
		}
	} else {
		should = []types.Query{
			constantTextMatch("text_german", textMatchBoost),
			constantTextMatch("text_german.unstemmed", unstemmedMatchBoost),
			constantMatch(exactMatchBoost, "text_german.keyword", "orth"),
		}
	}

	return types.Query{
		Bool: &types.BoolQuery{Should: should, MinimumShouldMatch: 1},
	}
}

// aggregationValuesName is the name of the terms aggregation nested in the
// filter aggregation of each requested field.
const aggregationValuesName = "values"

// maxAggregationBuckets is high enough to return every value of the supported
// fields (5 jlpt levels, ~25 grammar types, ~540 usage labels).
const maxAggregationBuckets = 600

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
	if len(filter.Usg) > 0 {
		values := make([]types.FieldValue, len(filter.Usg))
		for i, v := range filter.Usg {
			values[i] = v
		}
		filters[dictionary.EntryAggregationUsg] = termsQuery("usg", values)
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

// entryFilterFields are the fields with a field filter, in the order the
// filters are applied.
var entryFilterFields = []dictionary.EntryAggregationField{
	dictionary.EntryAggregationJlpt,
	dictionary.EntryAggregationGrammarType,
	dictionary.EntryAggregationUsg,
}

// queriesOf returns the filters in a stable order, leaving out the given field.
func queriesOf(filters map[dictionary.EntryAggregationField]types.Query, except dictionary.EntryAggregationField) []types.Query {
	var queries []types.Query
	for _, field := range entryFilterFields {
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
