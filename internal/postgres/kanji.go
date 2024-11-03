package postgres

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/nihongosensei/nihongosensei-go/internal/dictionary"
)

func (d *PostgresDictionary) SelectKanji(kanjiChar string) (*dictionary.Kanji, error) {
	var kanji dictionary.Kanji
	if _, err := d.Goqu.From("kanji").Where(goqu.Ex{"kanji": kanjiChar}).ScanStruct(&kanji); err != nil {
		return nil, err
	}

	return &kanji, nil
}

func (d *PostgresDictionary) SelectKanjis(kanjis []string) (*[]dictionary.Kanji, error) {
	k := make([]dictionary.Kanji, 0)
	if err := d.Goqu.From("kanji").Where(goqu.Ex{"kanji": kanjis}).ScanStructs(&k); err != nil {
		return nil, err
	}

	return &k, nil
}
