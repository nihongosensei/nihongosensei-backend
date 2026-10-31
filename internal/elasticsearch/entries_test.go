package elasticsearch

import (
	"encoding/json"
	"testing"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"

	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
)

func TestEntrySort(t *testing.T) {
	marshal := func(f dictionary.EntryFilter) string {
		b, err := json.Marshal(entrySort(f))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	if got := marshal(dictionary.EntryFilter{}); got != "null" {
		t.Errorf("expected no sort by default, got %s", got)
	}

	got := marshal(dictionary.EntryFilter{Sort: []dictionary.EntrySort{{Field: "jlpt", Order: "desc"}}})
	want := `[{"jlpt":{"order":"desc"}},{"id":{"order":"asc"}}]`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	got = marshal(dictionary.EntryFilter{Query: "x", Sort: []dictionary.EntrySort{{Field: "grammar_type"}}})
	want = `[{"grammar_type":{"order":"asc"}},{"_score":{"order":"desc"}},{"id":{"order":"asc"}}]`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestEntryAggregations(t *testing.T) {
	marshal := func(f dictionary.EntryFilter) string {
		b, err := json.Marshal(entryAggregations(f, entryFieldFilters(f)))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	if got := marshal(dictionary.EntryFilter{Jlpt: []int16{1}}); got != "null" {
		t.Errorf("expected no aggregations by default, got %s", got)
	}

	// each aggregation respects the other field's filter, but not its own
	f := dictionary.EntryFilter{
		Jlpt:         []int16{1, 2},
		GrammarType:  []string{"doushi"},
		Aggregations: []dictionary.EntryAggregationField{"jlpt", "grammar_type"},
	}
	want := `{"grammar_type":{"aggregations":{"values":{"terms":{"field":"grammar_type","size":50}}},"filter":{"bool":{"filter":[{"terms":{"jlpt":[1,2]}}]}}},` +
		`"jlpt":{"aggregations":{"values":{"terms":{"field":"jlpt","size":50}}},"filter":{"bool":{"filter":[{"terms":{"grammar_type":["doushi"]}}]}}}}`
	if got := marshal(f); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestParseEntryAggregations(t *testing.T) {
	filter := dictionary.EntryFilter{Aggregations: []dictionary.EntryAggregationField{"jlpt", "grammar_type"}}
	aggs := map[string]types.Aggregate{
		"jlpt": &types.FilterAggregate{Aggregations: map[string]types.Aggregate{
			"values": &types.LongTermsAggregate{Buckets: []types.LongTermsBucket{{Key: 3, DocCount: 7}}},
		}},
		"grammar_type": &types.FilterAggregate{Aggregations: map[string]types.Aggregate{
			"values": &types.StringTermsAggregate{Buckets: []types.StringTermsBucket{{Key: "doushi", DocCount: 2}}},
		}},
	}

	got, err := parseEntryAggregations(filter, aggs)
	if err != nil {
		t.Fatal(err)
	}
	if b := got["jlpt"]; len(b) != 1 || b[0] != (dictionary.AggregationBucket{Key: "3", Count: 7}) {
		t.Errorf("unexpected jlpt buckets %v", b)
	}
	if b := got["grammar_type"]; len(b) != 1 || b[0] != (dictionary.AggregationBucket{Key: "doushi", Count: 2}) {
		t.Errorf("unexpected grammar_type buckets %v", b)
	}

	if got, _ := parseEntryAggregations(dictionary.EntryFilter{}, nil); got != nil {
		t.Errorf("expected no aggregations, got %v", got)
	}
}

func TestEntryTextQuery(t *testing.T) {
	b, err := json.Marshal(entryTextQuery("Haus"))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"bool":{"minimum_should_match":1,"should":[` +
		`{"match":{"text_japanese":{"boost":2,"operator":"and","query":"Haus"}}},` +
		`{"match":{"text_german":{"boost":2,"operator":"and","query":"Haus"}}},` +
		`{"term":{"text_japanese.keyword":{"boost":30,"value":"haus"}}},` +
		`{"term":{"text_german.keyword":{"boost":30,"value":"haus"}}},` +
		`{"term":{"orth":{"boost":30,"value":"haus"}}},` +
		`{"term":{"reading":{"boost":30,"value":"haus"}}}]}}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}
