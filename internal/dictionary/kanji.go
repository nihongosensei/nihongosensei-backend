package dictionary

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/doug-martin/goqu/v9"
)

type Kanji struct {
	Kanji   string `json:"kanji"`
	Strokes int16  `json:"strokes"`
	Grade   int16  `json:"grade"`
	Kunyomi string `json:"kunyomi,omitempty"`
	Onyomi  string `json:"onyomi,omitempty"`
	Radical string `json:"radical,omitempty"`
}

func (d *Dictionary) SelectKanji(kanjiChar string) (Kanji, error) {
	var kanji Kanji
	if _, err := d.Goqu.From("kanji").Where(goqu.Ex{"kanji": kanjiChar}).ScanStruct(&kanji); err != nil {
		return Kanji{}, err
	}

	return kanji, nil
}

func (d *Dictionary) SelectKanjis(kanjiSet mapset.Set[string]) ([]Kanji, error) {
	kanjis := make([]Kanji, 0)
	if err := d.Goqu.From("kanji").Where(goqu.Ex{"kanji": kanjiSet.ToSlice()}).ScanStructs(&kanjis); err != nil {
		return kanjis, err
	}

	return kanjis, nil
}
