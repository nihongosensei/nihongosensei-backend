package dictionary

import (
	"context"
	"fmt"

	"github.com/guregu/null/v5"
)

var ErrEntryNotFound = fmt.Errorf("entry not found")
var ErrKanjiNotFound = fmt.Errorf("kanji not found")

type Dictionary interface {
	SelectEntry(ctx context.Context, id int) (*Entry, error)
	SelectEntries(ctx context.Context, filter EntryFilter) (*SelectEntries, error)
	SelectKanji(ctx context.Context, kanjiChar string) (*Kanji, error)
	SelectKanjis(ctx context.Context, kanjiSet KanjiFilters) (*SelectKanjis, error)
	SelectResolvedEntry(ctx context.Context, id int) (*ResolvedEntry, error)
}

type Kanji struct {
	Kanji   string `json:"kanji"`
	Strokes int16  `json:"strokes"`
	Grade   int16  `json:"grade"`
	Kunyomi string `json:"kunyomi,omitempty"`
	Onyomi  string `json:"onyomi,omitempty"`
	Radical string `json:"radical,omitempty"`
}

type KanjiFilters struct {
	Limit   int
	Offset  int
	Strokes []int16
	Grade   []int16
	Radical []string
	Kanjis  []string
}

type SelectKanjis struct {
	Kanjis []Kanji `json:"kanjis"`
	Count  int     `json:"count"`
}

type SelectEntries struct {
	Entries []Entry `json:"entries"`
	Count   int     `json:"count"`
}

type EntryFilter struct {
	Query  string
	Limit  int
	Offset int
	// The JLPT Level (1-5)
	Jlpt int16
	// A list of ids to select. Only existing matches are returned.
	Ids []int
	// A list of kanjis where at least one of them must match.
	Kanjis []string
	// A text which must exactly be contained in an entry.
	Text string
}

type Entry struct {
	Id        int        `json:"id"`
	EntryJson EntryJson  `json:"entryJson"`
	Jlpt      null.Int16 `json:"jlpt"`
	Kanjis    Kanjis     `json:"kanjis"`
}

type EntryJson struct {
	Id      int    `json:"id"`
	Expl    []any  `json:"expl,omitempty"`
	Form    Form   `json:"form"`
	Usg     []Usg  `json:"usg,omitempty"`
	GramGrp any    `json:"gramGrp,omitempty"`
	Sense   []any  `json:"sense,omitempty"`
	Ruigos  Ruigos `json:"ruigos"`
	Ref     []any  `json:"ref,omitempty"`
}

type ResolvedEntry struct {
	Entry     *Entry                       `json:"entry"`
	Synonyms  []Entry                      `json:"synonyms,omitempty"`
	Relations []EntryRefWithResolvedSource `json:"relatedEntries,omitempty"`
	Kanjis    []Kanji                      `json:"kanjis,omitempty"`
}

type EntryRefWithResolvedSource struct {
	Type         string    `json:"type"`
	SubEntryType string    `json:"subentrytype,omitempty"`
	EntryJson    EntryJson `json:"entry"`
}

type Synonyms []int

type Kanjis []string

type Form struct {
	Orth    []Orth  `json:"orth,omitempty"`
	Reading Reading `json:"reading"`
}

type Orth struct {
	Value     string `json:"value"`
	Midashigo bool   `json:"midashigo"`
	Type      string `json:"type,omitempty"`
}

type Reading struct {
	Hira    string   `json:"hira,omitempty"`
	Hatsuon string   `json:"hatsuon,omitempty"`
	Accent  []string `json:"accent,omitempty"`
}

type Usg struct {
	Type    string `json:"type,omitempty"`
	Reg     string `json:"reg,omitempty"`
	Content string `json:"content"`
}

type Ruigos struct {
	Ruigo []Ruigo `json:"ruigo,omitempty"`
}

type Ruigo struct {
	Id int `json:"id"`
}
