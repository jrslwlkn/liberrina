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
        added_at,
        user_id
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
        datetime(),
        @user_id
    ) returning lang_id;

-- name: AddDoc :one
insert into
    docs(
        title,
        author,
        body,
        notes,
        lang_id,
        user_id,
        added_at,
        term_count,
        sentence_count,
        terms_new
    )
values
    (
        @title,
        @author,
        @body,
        @notes,
        @lang_id,
        @user_id,
        datetime(),
        0,
        0,
        0
    ) returning doc_id;

-- name: InitDoc :exec
update
    docs
set
    term_count = @term_count,
    sentence_count = @sentence_count
where
    doc_id = @doc_id;

-- name: PruneChunks :exec
delete from
    chunks
where
    doc_id = @doc_id;

-- name: AddChunk :exec
insert into
    chunks (
        doc_id,
        position,
        value,
        suffix
    )
values
    (@doc_id, @position, @value, @suffix);

-- name: AddTerms :exec
insert into
    terms(
        value,
        translation,
        term_level_id,
        lang_id,
        user_id,
        added_at
    )
select
    value,
    '',
    1,
    d.lang_id,
    d.user_id,
    datetime()
from
    chunks c
    join docs d on c.doc_id = d.doc_id
where
    d.doc_id = @doc_id
    and not exists (
        select
            value
        from
            terms
        where
            user_id = (
                select
                    user_id
                from
                    docs
                where
                    doc_id = @doc_id
            )
    )
group by
    value;
