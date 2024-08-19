-- name: GetDocs :many
select
    doc_id,
    title,
    author,
    added_at,
    term_count,
    terms_new,
    sentence_count
from
    docs;

-- name: GetAllLangs :many
select
    id,
    name
from
    langs_dim;

-- name: GetLangs :many
select
    lang_id,
    name
from
    langs;

-- name: GetDocByID :one
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
    doc_id = @id;

-- name: AddLang :one
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
        @name,
        @from_id,
        @to_id,
        @quick_lookup_URI,
        @lookup_URI_1,
        @lookup_URI_2,
        @chars_pattern,
        @sentence_sep,
        datetime()
    ) returning *;

-- name: PruneChunks :exec
delete from
    chunks
where
    doc_id = @doc_id;

-- name: AddChunk :one
insert into
    chunks (
        doc_id,
        position,
        value,
        suffix
    )
values
    (@doc_id, @position, @value, @suffix) returning *;

    
