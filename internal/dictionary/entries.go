package dictionary

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type EntryFilter struct {
	Query  string
	Limit  uint
	Offset uint
	Jlpt   int16
}

func (d *Dictionary) SelectEntries(filter EntryFilter) ([]Entry, error) {
	query := d.Goqu.From("entry").
		Select(Entry{}).
		Limit(filter.Limit).
		Offset(filter.Offset)

	var where []exp.Expression
	if filter.Jlpt != 0 {
		where = append(where, goqu.Ex{
			"jlpt": filter.Jlpt,
		})
	}
	if filter.Query != "" {
		query = query.From("entry", goqu.Func("plainto_tsquery", filter.Query).As("query"))
		where = append(where, goqu.L("textjson_tsvector @@ query"))
		// TODO: prefer entries which have an exact match
		// order by text search rank
		query = query.
			Order(goqu.Func("ts_rank", goqu.I("textjson_tsvector"), goqu.I("query")).Desc())
	}

	if len(where) > 0 {
		query = query.Where(where...)
	}

	entries := make([]Entry, 0)
	if err := query.ScanStructs(&entries); err != nil {
		return entries, err
	}

	return entries, nil
}

func (d *Dictionary) SelectEntriesCount(filter EntryFilter) (int, error) {
	query := d.Goqu.From("entry").
		Select(goqu.COUNT("*"))

	var where []exp.Expression
	if filter.Jlpt != 0 {
		where = append(where, goqu.Ex{
			"jlpt": filter.Jlpt,
		})
	}
	if filter.Query != "" {
		query = query.From("entry", goqu.Func("plainto_tsquery", filter.Query).As("query"))
		where = append(where, goqu.L("textjson_tsvector @@ query"))
	}

	if len(where) > 0 {
		query = query.Where(where...)
	}

	count := 0
	_, err := query.ScanVal(&count)

	return count, err
}
