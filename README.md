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
