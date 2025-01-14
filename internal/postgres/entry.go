package postgres

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
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
	if err := d.Goqu.From("entry").Where(goqu.Ex{"id": ids}).ScanStructs(&entries); err != nil {
		return nil, err
	}

	return &entries, nil
}
