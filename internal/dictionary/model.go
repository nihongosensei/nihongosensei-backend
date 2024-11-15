package dictionary

import (
	"encoding/json"
	"errors"
	"github.com/guregu/null/v5"
)

type Dictionary interface {
	SelectEntry(id int) (*Entry, error)
	SelectEntryByText(text string) (*Entry, error)
	SelectEntries(filter EntryFilter) (*SelectEntries, error)
	SelectEntryByIds(ids []int) (*[]Entry, error)
	SelectMainRefsByTargetId(id int) (*[]EntryRefWithResolvedSource, error)
	SelectKanji(kanjiChar string) (*Kanji, error)
	SelectKanjis(kanjiSet []string) (*[]Kanji, error)
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
	Ids []int16
	// A list of kanjis where at least one of them must match.
	Kanjis []string
}

type Entry struct {
	Id        int        `json:"id" db:"id"`
	EntryJson EntryJson  `json:"entryJson" db:"entry_json"`
	Jlpt      null.Int16 `json:"jlpt,omitempty" db:"jlpt"`
	Kanjis    Kanjis     `json:"kanjis" db:"kanjis"`
	Synonyms  Synonyms   `json:"synonyms" db:"synonyms"`
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
	EntryJson    EntryJson `json:"entry" db:"entry_json"`
}

type Synonyms []int

// Scan on Synonyms allows decoding of embedded json
// TODO: move to postgres package?
func (p *Synonyms) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	return json.Unmarshal(source, p)
}

type Kanjis []string

// Scan on Kanjis allows decoding of embedded json
// TODO: move to postgres package?
func (p *Kanjis) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	return json.Unmarshal(source, p)
}

// Scan on EntryJson allows decoding of embedded json
// TODO: move to postgres package?
func (p *EntryJson) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	return json.Unmarshal(source, p)
}

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
