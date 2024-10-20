package dictionary

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/doug-martin/goqu/v9"
)

func (d *Dictionary) SelectEntry(id int) (Entry, error) {
	var entry Entry
	_, err := d.Goqu.From("entry").
		Select("id", "entryjson", "jlpt").
		Where(goqu.Ex{"id": id}).ScanStruct(&entry)
	if err != nil {
		return entry, err
	}

	return entry, nil
}

func (d *Dictionary) SelectEntryByIds(ids mapset.Set[int]) ([]Entry, error) {
	entries := make([]Entry, 0)
	err := d.Goqu.From("entry").
		Where(goqu.Ex{"id": ids.ToSlice()}).
		ScanStructs(&entries)
	if err != nil {
		return entries, err
	}

	return entries, nil
}

func (d *Dictionary) SelectEntryByText(text string) (Entry, error) {
	var entry Entry
	// necessary as "?" json operator can not be used with goqu
	err := d.DB.QueryRow("SELECT id, entryjson, jlpt FROM entry WHERE textjson ? $1", text).
		Scan(&entry.Id, &entry.EntryJson, &entry.Jlpt)

	return entry, err
}
