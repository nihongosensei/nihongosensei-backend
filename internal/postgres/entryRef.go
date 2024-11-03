package postgres

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/nihongosensei/nihongosensei-go/internal/dictionary"
)

func (d *PostgresDictionary) SelectMainRefsByTargetId(id int) (*[]dictionary.EntryRefWithResolvedSource, error) {
	query := d.Goqu.From("entry_ref").
		Select("type", "subentrytype", "entry.entry_json").
		Join(
			goqu.I("entry"),
			goqu.On(goqu.Ex{"entry.id": goqu.I("entry_ref.source_id")}),
		).
		Where(goqu.Ex{
			"entry_ref.target_id": id,
			"entry_ref.type":      "MAIN",
		})

	entries := make([]dictionary.EntryRefWithResolvedSource, 0)
	if err := query.ScanStructs(&entries); err != nil {
		return nil, err
	}

	return &entries, nil
}
