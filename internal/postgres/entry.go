package postgres

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/nihongosensei/nihongosensei-go/internal/dictionary"
)

func (d *PostgresDictionary) SelectEntry(id int) (*dictionary.Entry, error) {
	var entry dictionary.Entry
	_, err := d.Goqu.From("entry").
		Select(dictionary.Entry{}).
		Where(goqu.Ex{"id": id}).ScanStruct(&entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (d *PostgresDictionary) SelectEntryByIds(ids []int) (*[]dictionary.Entry, error) {
	entries := make([]dictionary.Entry, 0)
	err := d.Goqu.From("entry").
		Where(goqu.Ex{"id": ids}).
		ScanStructs(&entries)
	if err != nil {
		return nil, err
	}

	return &entries, nil
}

func (d *PostgresDictionary) SelectEntryByText(text string) (*dictionary.Entry, error) {
	var entry dictionary.Entry
	// necessary as "?" json operator can not be used with goqu
	err := d.DB.QueryRow("SELECT id, entry_json, jlpt FROM entry WHERE entry.text ? $1", text).
		Scan(&entry.Id, &entry.EntryJson, &entry.Jlpt)

	return &entry, err
}
