package dictionary

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/doug-martin/goqu/v9"
	"github.com/guregu/null/v5"
)

type Dictionary struct {
	DB   *sql.DB
	Goqu *goqu.Database
}

type Entry struct {
	Id        int        `json:"id" db:"id"`
	EntryJson EntryJson  `json:"entry_json" db:"entryjson"`
	Jlpt      null.Int16 `json:"jlpt,omitempty" db:"jlpt"`
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

// Scan on EntryJson allows decoding of embedded json
func (p *EntryJson) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	err := json.Unmarshal(source, p)
	if err != nil {
		return err
	}

	return nil
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
