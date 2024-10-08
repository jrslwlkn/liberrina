// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: queries.sql

package queries

import (
	"context"
	"database/sql"
	"time"
)

const addChunk = `-- name: AddChunk :one
insert into
    chunks (
        doc_id,
        position,
        value,
        suffix
    )
values
    (?1, ?2, ?3, ?4) returning doc_id, position, value, suffix
`

type AddChunkParams struct {
	DocID    int64
	Position int64
	Value    interface{}
	Suffix   interface{}
}

func (q *Queries) AddChunk(ctx context.Context, arg AddChunkParams) (Chunk, error) {
	row := q.db.QueryRowContext(ctx, addChunk,
		arg.DocID,
		arg.Position,
		arg.Value,
		arg.Suffix,
	)
	var i Chunk
	err := row.Scan(
		&i.DocID,
		&i.Position,
		&i.Value,
		&i.Suffix,
	)
	return i, err
}

const addLang = `-- name: AddLang :one
insert into
    langs(
        name,
        from_id,
        to_id,
        quick_lookup_uri,
        lookup_uri_1,
        lookup_uri_2,
        chars_pattern,
        sentence_sep,
        added_at
    )
values
    (
        ?1,
        ?2,
        ?3,
        ?4,
        ?5,
        ?6,
        ?7,
        ?8,
        datetime()
    ) returning lang_id, name, from_id, to_id, quick_lookup_uri, lookup_uri_1, lookup_uri_2, chars_pattern, sentence_sep, user_id, added_at
`

type AddLangParams struct {
	Name           string
	FromID         string
	ToID           string
	QuickLookupURI string
	LookupURI1     sql.NullString
	LookupURI2     sql.NullString
	CharsPattern   string
	SentenceSep    string
}

func (q *Queries) AddLang(ctx context.Context, arg AddLangParams) (Lang, error) {
	row := q.db.QueryRowContext(ctx, addLang,
		arg.Name,
		arg.FromID,
		arg.ToID,
		arg.QuickLookupURI,
		arg.LookupURI1,
		arg.LookupURI2,
		arg.CharsPattern,
		arg.SentenceSep,
	)
	var i Lang
	err := row.Scan(
		&i.LangID,
		&i.Name,
		&i.FromID,
		&i.ToID,
		&i.QuickLookupUri,
		&i.LookupUri1,
		&i.LookupUri2,
		&i.CharsPattern,
		&i.SentenceSep,
		&i.UserID,
		&i.AddedAt,
	)
	return i, err
}

const getAllLangs = `-- name: GetAllLangs :many
select
    id,
    name
from
    langs_dim
`

func (q *Queries) GetAllLangs(ctx context.Context) ([]LangsDim, error) {
	rows, err := q.db.QueryContext(ctx, getAllLangs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LangsDim
	for rows.Next() {
		var i LangsDim
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDocByID = `-- name: GetDocByID :one
select
    doc_id,
    title,
    author,
    added_at,
    term_count,
    terms_new,
    sentence_count
from
    docs
where
    doc_id = ?1
`

type GetDocByIDRow struct {
	DocID         int64
	Title         string
	Author        sql.NullString
	AddedAt       time.Time
	TermCount     int64
	TermsNew      int64
	SentenceCount int64
}

func (q *Queries) GetDocByID(ctx context.Context, id int64) (GetDocByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getDocByID, id)
	var i GetDocByIDRow
	err := row.Scan(
		&i.DocID,
		&i.Title,
		&i.Author,
		&i.AddedAt,
		&i.TermCount,
		&i.TermsNew,
		&i.SentenceCount,
	)
	return i, err
}

const getDocs = `-- name: GetDocs :many
select
    doc_id,
    title,
    author,
    added_at,
    term_count,
    terms_new,
    sentence_count
from
    docs
`

type GetDocsRow struct {
	DocID         int64
	Title         string
	Author        sql.NullString
	AddedAt       time.Time
	TermCount     int64
	TermsNew      int64
	SentenceCount int64
}

func (q *Queries) GetDocs(ctx context.Context) ([]GetDocsRow, error) {
	rows, err := q.db.QueryContext(ctx, getDocs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDocsRow
	for rows.Next() {
		var i GetDocsRow
		if err := rows.Scan(
			&i.DocID,
			&i.Title,
			&i.Author,
			&i.AddedAt,
			&i.TermCount,
			&i.TermsNew,
			&i.SentenceCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLangs = `-- name: GetLangs :many
select
    lang_id,
    name
from
    langs
`

type GetLangsRow struct {
	LangID int64
	Name   string
}

func (q *Queries) GetLangs(ctx context.Context) ([]GetLangsRow, error) {
	rows, err := q.db.QueryContext(ctx, getLangs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetLangsRow
	for rows.Next() {
		var i GetLangsRow
		if err := rows.Scan(&i.LangID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const pruneChunks = `-- name: PruneChunks :exec
delete from
    chunks
where
    doc_id = ?1
`

func (q *Queries) PruneChunks(ctx context.Context, docID int64) error {
	_, err := q.db.ExecContext(ctx, pruneChunks, docID)
	return err
}
