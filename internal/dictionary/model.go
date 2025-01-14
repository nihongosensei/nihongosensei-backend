package dictionary

import (
	"fmt"
	"github.com/guregu/null/v5"
)

var ErrEntryNotFound = fmt.Errorf("entry not found")
var ErrKanjiNotFound = fmt.Errorf("kanji not found")

type Dictionary interface {
	SelectEntry(id int) (*Entry, error)
	SelectEntries(filter EntryFilter) (*SelectEntries, error)
	SelectKanji(kanjiChar string) (*Kanji, error)
	SelectKanjis(kanjiSet []string) (*[]Kanji, error)
	SelectResolvedEntry(id int) (*ResolvedEntry, error)
}

type Kanji struct {
	Kanji   string `json:"kanji"`
	Strokes int16  `json:"strokes"`
	Grade   int16  `json:"grade"`
	Kunyomi string `json:"kunyomi,omitempty"`
	Onyomi  string `json:"onyomi,omitempty"`
	Radical string `json:"radical,omitempty"`
}

type SelectEntries struct {
	Entries []Entry `json:"entries"`
	Count   int     `json:"count"`
}

type EntryFilter struct {
	Query  string
	Limit  uint
	Offset uint
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
	Jlpt      null.Int16 `json:"jlpt,omitempty"`
	Kanjis    Kanjis     `json:"kanjis"`
}

type EntryJson struct {
	Id      int    `json:"id"`
	Expl    []any  `json:"expl,omitempty"`
	Form    Form   `json:"form,omitempty"`
	Usg     []Usg  `json:"usg,omitempty"`
	GramGrp any    `json:"gramGrp,omitempty"`
	Sense   []any  `json:"sense,omitempty"`
	Ruigos  Ruigos `json:"ruigos,omitempty"`
	Ref     []any  `json:"ref,omitempty"`
}

type ResolvedEntry struct {
	Entry     *Entry                        `json:"entry"`
	Synonyms  *[]Entry                      `json:"synonyms"`
	Relations *[]EntryRefWithResolvedSource `json:"relatedEntries"`
	Kanjis    *[]Kanji                      `json:"kanjis"`
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
	Reading Reading `json:"reading,omitempty"`
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
