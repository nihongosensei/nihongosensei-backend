package postgres

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/nihongosensei/nihongosensei-go/internal/dictionary"
	"log/slog"
)

func (d *PostgresDictionary) SelectEntries(filter dictionary.EntryFilter) (*dictionary.SelectEntries, error) {
	query := d.Goqu.From("entry").
		Select(dictionary.Entry{}).
		Limit(filter.Limit).
		Offset(filter.Offset)

	var where []exp.Expression
	if filter.Jlpt != 0 {
		where = append(where, goqu.Ex{
			"jlpt": filter.Jlpt,
		})
	}
	if filter.Query != "" {
		query = query.From(
			"entry",
			goqu.Func("plainto_tsquery", filter.Query).As("query"),
			goqu.Func("to_jsonb", goqu.L("?::text", filter.Query)).As("queryjson"),
		)
		where = append(where, goqu.L("text_tsv @@ query"))
		// order by text search rank and prefer entries that have an exact match
		query = query.
			Order(
				goqu.Func(
					"greatest",
					goqu.Func(
						"ts_rank", goqu.I("text_tsv"),
						goqu.I("query"),
					),
					goqu.L("CASE WHEN (text @> queryjson) THEN 1 ELSE 0 END"),
				).Desc(),
			)
	}
	if len(filter.Ids) > 0 {
		where = append(where, goqu.Ex{
			"id": filter.Ids,
		})
	}
	if len(filter.Kanjis) > 0 {
		or := make([]goqu.Expression, len(filter.Kanjis))
		for i, k := range filter.Kanjis {
			or[i] = goqu.L("kanjis @> ?::jsonb", fmt.Sprintf(`["%s"]`, k))
		}
		where = append(where, goqu.Or(or...))
	}

	if len(where) > 0 {
		query = query.Where(where...)
	}

	sql, _, _ := query.ToSQL()
	slog.Debug("created entries query", "query", sql)

	entries := make([]dictionary.Entry, 0)
	if err := query.ScanStructs(&entries); err != nil {
		return nil, err
	}

	count, err := d.selectEntriesCount(filter)
	if err != nil {
		return nil, err
	}

	return &dictionary.SelectEntries{
		Entries: entries,
		Count:   count,
	}, nil
}

func (d *PostgresDictionary) selectEntriesCount(filter dictionary.EntryFilter) (int, error) {
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
		where = append(where, goqu.L("text_tsv @@ query"))
	}
	if len(filter.Ids) > 0 {
		where = append(where, goqu.Ex{
			"id": filter.Ids,
		})
	}
	if len(filter.Kanjis) > 0 {
		or := make([]goqu.Expression, len(filter.Kanjis))
		for i, k := range filter.Kanjis {
			or[i] = goqu.L("kanjis @> ?::jsonb", fmt.Sprintf(`["%s"]`, k))
		}
		where = append(where, goqu.Or(or...))
	}

	if len(where) > 0 {
		query = query.Where(where...)
	}

	sql, _, _ := query.ToSQL()
	slog.Debug("created entries count query", "query", sql)

	count := 0
	_, err := query.ScanVal(&count)

	return count, err
}
