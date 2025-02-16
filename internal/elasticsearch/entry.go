package elasticsearch

import (
	"context"
	"encoding/json"
	"github.com/guregu/null/v5"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
	"log/slog"
	"strconv"
)

type ElasticEntry struct {
	Id        int                  `json:"id"`
	EntryJson dictionary.EntryJson `json:"entryJson"`
	Jlpt      null.Int16           `json:"jlpt,omitempty"`
	Kanjis    dictionary.Kanjis    `json:"kanjis"`
	Synonyms  dictionary.Synonyms  `json:"synonyms"`
	Refs      []EntryRef           `json:"refs"`
}

func (e *ElasticEntry) toEntry() *dictionary.Entry {
	return &dictionary.Entry{
		Id:        e.Id,
		EntryJson: e.EntryJson,
		Jlpt:      e.Jlpt,
		Kanjis:    e.Kanjis,
	}
}

type EntryRef struct {
	Type         string `json:"type"`
	SubEntryType string `json:"subentrytype,omitempty"`
	TargetId     int    `json:"targetId"`
}

func (d *ElasticsearchDictionary) SelectEntry(id int) (*dictionary.Entry, error) {
	entry, err := d.selectElasticEntry(id)
	if err != nil {
		return nil, err
	}

	return entry.toEntry(), nil
}

func (d *ElasticsearchDictionary) selectElasticEntry(id int) (*ElasticEntry, error) {
	res, err := d.TypedClient.Core.Get("entries", strconv.Itoa(id)).Do(context.Background())

	if err != nil {
		return nil, err
	}
	if !res.Found {
		return nil, dictionary.ErrEntryNotFound
	}

	var entry ElasticEntry
	err = json.Unmarshal(res.Source_, &entry)

	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (d *ElasticsearchDictionary) selectEntryByIds(ids []int) ([]dictionary.Entry, error) {
	res, err := d.SelectEntries(dictionary.EntryFilter{
		Ids:   ids,
		Limit: len(ids),
	})

	if err != nil {
		return nil, err
	}

	return res.Entries, nil
}

func (d *ElasticsearchDictionary) selectResolvedRefs(entry *ElasticEntry) (*[]dictionary.EntryRefWithResolvedSource, error) {
	entries := make([]dictionary.EntryRefWithResolvedSource, 0)

	if len(entry.Refs) == 0 {
		return &entries, nil
	}

	entryRefIds := make([]int, len(entry.Refs))
	for i, ref := range entry.Refs {
		entryRefIds[i] = ref.TargetId
	}

	resp, err := d.selectEntryByIds(entryRefIds)
	if err != nil {
		return nil, err
	}

	for _, ref := range entry.Refs {
		var refEntry *dictionary.Entry = nil

		for _, entry := range resp {
			if entry.Id == ref.TargetId {
				refEntry = &entry
			}
		}

		if refEntry == nil {
			slog.Warn("failed to resolve reference entry", "entryId", entry.Id, "refTargetId", ref.TargetId)
			continue
		}

		entries = append(entries, dictionary.EntryRefWithResolvedSource{
			Type:         ref.Type,
			SubEntryType: ref.SubEntryType,
			EntryJson:    refEntry.EntryJson,
		})
	}

	return &entries, nil
}

func (d *ElasticsearchDictionary) SelectResolvedEntry(id int) (*dictionary.ResolvedEntry, error) {
	entry, err := d.selectElasticEntry(id)
	if err != nil {
		return nil, err
	}

	resolvedEntry := dictionary.ResolvedEntry{
		Entry: entry.toEntry(),
	}

	// Select Synonyms
	if entry.Synonyms != nil && len(entry.Synonyms) > 0 {
		sResult, sErr := d.SelectEntries(dictionary.EntryFilter{
			Ids: entry.Synonyms,
		})
		if sErr != nil {
			return nil, sErr
		}

		resolvedEntry.Synonyms = &sResult.Entries
	}

	// Select Kanjis
	if entry.Kanjis != nil && len(entry.Kanjis) > 0 {
		kanjis, kErr := d.SelectKanjis(dictionary.KanjiFilters{
			Kanjis: entry.Kanjis,
			Limit:  len(entry.Kanjis),
		})
		if kErr != nil {
			return nil, kErr
		}

		resolvedEntry.Kanjis = &kanjis.Kanjis
	}

	// Select References
	refs, err := d.selectResolvedRefs(entry)

	if err != nil {
		return nil, err
	}

	resolvedEntry.Relations = refs

	return &resolvedEntry, nil
}
