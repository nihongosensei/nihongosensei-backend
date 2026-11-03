package handler

import (
	"encoding/json"
	"testing"

	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
)

func TestEntriesByFilterSortDecoding(t *testing.T) {
	cases := map[string]bool{
		`{"sort":[{"field":"jlpt","order":"desc"}]}`: true,
		`{"sort":[{"field":"grammar_type"}]}`:        true,
		`{"sort":[{"field":"strokes"}]}`:             false,
		`{"sort":[{"field":"id","order":"up"}]}`:     false,
		`{"jlpt":[1,2],"grammar_type":["doushi"]}`:   true,
		`{"jlpt":3}`: false,
		`{"aggregations":["jlpt","grammar_type"]}`:       true,
		`{"aggregations":["strokes"]}`:                   false,
		`{"usg":["Med.","Bot."],"aggregations":["usg"]}`: true,
		`{"usg":"Med."}`:                                 false,
	}
	for body, ok := range cases {
		var input EntriesByFilterPostBody
		err := json.Unmarshal([]byte(body), &input)
		if (err == nil) != ok {
			t.Errorf("%s: expected ok=%v, got err=%v", body, ok, err)
		}
	}
}

func TestValidateSort(t *testing.T) {
	s := func(fields ...dictionary.EntrySortField) []dictionary.EntrySort {
		var out []dictionary.EntrySort
		for _, f := range fields {
			out = append(out, dictionary.EntrySort{Field: f})
		}
		return out
	}

	t.Run("ScoreRequiresQuery", func(t *testing.T) {
		if validateSort(s(dictionary.EntrySortScore), false) {
			t.Error("expected _score without query to be invalid")
		}
		if !validateSort(s(dictionary.EntrySortScore), true) {
			t.Error("expected _score with query to be valid")
		}
	})
	t.Run("RejectsDuplicates", func(t *testing.T) {
		if validateSort(s(dictionary.EntrySortId, dictionary.EntrySortId), false) {
			t.Error("expected duplicate fields to be invalid")
		}
	})
	t.Run("RejectsTooMany", func(t *testing.T) {
		if validateSort(s("a", "b", "c", "d"), true) {
			t.Error("expected too many criteria to be invalid")
		}
	})
}

func TestValidateAggregations(t *testing.T) {
	if !validateAggregations([]dictionary.EntryAggregationField{"jlpt", "grammar_type"}) {
		t.Error("expected distinct aggregations to be valid")
	}
	if validateAggregations([]dictionary.EntryAggregationField{"jlpt", "jlpt"}) {
		t.Error("expected duplicate aggregations to be invalid")
	}
}

func TestValidateJlpt(t *testing.T) {
	if !validateJlpt([]int16{1, 5}) {
		t.Error("expected levels 1-5 to be valid")
	}
	if validateJlpt([]int16{0}) || validateJlpt([]int16{6}) {
		t.Error("expected out-of-range levels to be invalid")
	}
}
