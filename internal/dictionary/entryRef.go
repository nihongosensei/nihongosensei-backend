package dictionary

import "github.com/doug-martin/goqu/v9"

type EntryRefWithResolvedSource struct {
	Type         string    `json:"type"`
	SubEntryType string    `json:"subentrytype,omitempty"`
	EntryJson    EntryJson `json:"entry"`
}

func (d *Dictionary) SelectMainRefsByTargetId(id int) ([]EntryRefWithResolvedSource, error) {
	query := d.Goqu.From("entry_ref").
		Select("type", "subentrytype", "entry.entryjson").
		Join(
			goqu.I("entry"),
			goqu.On(goqu.Ex{"entry.id": goqu.I("entry_ref.source_id")}),
		).
		Where(goqu.Ex{
			"entry_ref.target_id": id,
			"entry_ref.type":      "MAIN",
		})

	entries := make([]EntryRefWithResolvedSource, 0)
	if err := query.ScanStructs(&entries); err != nil {
		return entries, err
	}

	return entries, nil
}
