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
	want := `{"grammar_type":{"aggregations":{"values":{"terms":{"field":"grammar_type","size":600}}},"filter":{"bool":{"filter":[{"terms":{"jlpt":[1,2]}}]}}},` +
		`"jlpt":{"aggregations":{"values":{"terms":{"field":"jlpt","size":600}}},"filter":{"bool":{"filter":[{"terms":{"grammar_type":["doushi"]}}]}}}}`
	if got := marshal(f); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestEntryAggregationsUsg(t *testing.T) {
	f := dictionary.EntryFilter{
		GrammarType:  []string{"meishi"},
		Usg:          []string{"Med.", "Bot."},
		Aggregations: []dictionary.EntryAggregationField{"usg"},
	}
	b, err := json.Marshal(entryAggregations(f, entryFieldFilters(f)))
	if err != nil {
		t.Fatal(err)
	}
	// the usg aggregation ignores the usg filter, but respects the grammar_type one
	want := `{"usg":{"aggregations":{"values":{"terms":{"field":"usg","size":600}}},"filter":{"bool":{"filter":[{"terms":{"grammar_type":["meishi"]}}]}}}}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}

	b, err = json.Marshal(queriesOf(entryFieldFilters(f), ""))
	if err != nil {
		t.Fatal(err)
	}
	want = `[{"terms":{"grammar_type":["meishi"]}},{"terms":{"usg":["Med.","Bot."]}}]`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
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
	marshal := func(text string) string {
		b, err := json.Marshal(entryTextQuery(text))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	want := `{"bool":{"minimum_should_match":1,"should":[` +
		`{"constant_score":{"boost":2,"filter":{"match":{"text_german":{"operator":"and","query":"Haus"}}}}},` +
		`{"constant_score":{"boost":1,"filter":{"match":{"text_german.unstemmed":{"operator":"and","query":"Haus"}}}}},` +
		`{"constant_score":{"boost":30,"filter":{"bool":{"minimum_should_match":1,"should":[` +
		`{"match":{"text_german.keyword":{"query":"Haus"}}},` +
		`{"match":{"orth":{"query":"Haus"}}}]}}}}]}}`
	if got := marshal("Haus"); got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = `{"bool":{"minimum_should_match":1,"should":[` +
		`{"match":{"text_japanese":{"boost":2,"operator":"and","query":"食べる"}}},` +
		`{"constant_score":{"boost":5,"filter":{"bool":{"minimum_should_match":1,"should":[` +
		`{"match":{"reading.prefix":{"query":"食べる"}}},` +
		`{"match":{"orth.prefix":{"query":"食べる"}}}]}}}},` +
		`{"constant_score":{"boost":30,"filter":{"bool":{"minimum_should_match":1,"should":[` +
		`{"match":{"text_japanese.keyword":{"query":"食べる"}}},` +
		`{"match":{"orth":{"query":"食べる"}}},` +
		`{"match":{"reading":{"query":"食べる"}}}]}}}}]}}`
	if got := marshal("食べる"); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestFrequencyQuery(t *testing.T) {
	b, err := json.Marshal(frequencyQuery())
	if err != nil {
		t.Fatal(err)
	}
	want := `{"rank_feature":{"boost":10,"field":"frequency","saturation":{"pivot":0.0002}}}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}
