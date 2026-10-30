package elasticsearch

import (
	"encoding/json"
	"testing"

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
