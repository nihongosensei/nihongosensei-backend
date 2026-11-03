# 日本語先生 (Nihongosensei) Backend

This project provides an api to japanese related data written in go.

The data is stored via elasticsearch. For importing data,
see [wadoku-export-reader](https://github.com/nihongosensei/wadoku-export-reader).

Currently, the following endpoints are available:

| URL                           | Method | Parameters                    | Description                                                                          |
|-------------------------------|--------|-------------------------------|--------------------------------------------------------------------------------------|
| `/api/furigana`               | POST   | `{text: string}`              | Generates furigana for a given text.                                                 |
| `/api/entry/id/{id}`          | GET    | -                             | Returns a dictionary entry.                                                          |
| `/api/entry/id/{id}/resolved` | GET    | -                             | Returns a dictionary entry with resolved relations (e.g. synonyms, related entries). |
| `/api/entries`                | POST   | See `EntriesByFilterPostBody` | Returns a list of entries.                                                           |
| `/api/kanji/{kanji}`          | GET    | -                             | Returns a kanji entry.                                                               |
| `/api/kanjis`                 | POST   | See `KanjisByFilterPostBody`  | Returns a list of kanjis.                                                            |
## Sorting

`/api/entries` accepts an optional `sort` array in the request body. Criteria are applied in the given order:

```json
{ "query": "食べる", "sort": [{ "field": "jlpt", "order": "desc" }] }
```

| Field          | Notes                                                       |
|----------------|-------------------------------------------------------------|
| `id`           |                                                             |
| `jlpt`         | Entries without a JLPT level are always sorted last.        |
| `grammar_type` |                                                             |
| `_score`       | Relevance. Only allowed when `query` is set.                |

`order` is `asc` (default) or `desc` (default for `_score`). Without `sort`, entries are ordered by relevance.
At most 3 criteria are allowed and each field may only appear once, otherwise the request is rejected with `400`.
Ties are broken by relevance (when searching) and then `id`, so paging is stable.

## Filtering and aggregations

`/api/entries` accepts the list filters `jlpt` (levels 1-5), `grammar_type` and `usg` (usage labels like `Med.` or `Bot.`, case-sensitive). Values within a filter are combined with OR, different filters with AND:

```json
{ "jlpt": [1, 2], "grammar_type": ["doushi", "meishi"], "usg": ["Med."], "aggregations": ["jlpt", "grammar_type", "usg"] }
```

Entries which have a JLPT level of 1 or 2 **and** are a doushi or a meishi **and** are labeled `Med.` are returned. Out-of-range levels are rejected with `400`.

> **Breaking change:** `jlpt` used to be a single number and must now be an array.

`aggregations` is optional and may contain `jlpt`, `grammar_type` and `usg`. The response then contains bucket counts per value, ordered by count:

```json
{ "entries": [], "count": 749, "aggregations": { "jlpt": [{ "key": "1", "count": 453 }] } }
```

Aggregations are not affected by `limit` and `offset`. Each aggregation ignores the filter on its own field but respects the query and all other filters, so unselected values of a field stay visible (faceted search).
Entries without a value are not counted. Unknown or duplicate aggregation fields are rejected with `400`.
