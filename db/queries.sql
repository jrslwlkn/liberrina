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
    docs
where
    user_id = @user_id;

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
    langs
where
    user_id = @user_id;

-- name: GetDocMeta :one
select
    doc_id,
    title,
    author,
    notes,
    d.added_at,
    term_count,
    sentence_count,
    terms_new,
    from_id from_lang_id,
    to_id to_lang_id,
    quick_lookup_uri,
    lookup_uri_1,
    lookup_uri_2,
    chars_pattern,
    sentence_sep
from
    docs d
    join langs l on d.lang_id = l.lang_id
where
    doc_id = @id;

-- name: GetDocBody :many
select
    c.value,
    c.suffix,
    case when t.term_level_id is null then 1 else t.term_level_id end as term_level_id,
    case when t.translation is null then '' else t.translation end as translation
from
    chunks c
    left join terms t 
        on c.value = t.value and t.user_id = @user_id
where
    c.doc_id = @doc_id;

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
        @notes,
        @lang_id,
        @user_id,
        datetime(),
        0,
        0,
        0
    ) returning doc_id;

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
    c.value,
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
    and c.value not in (
        select
            value
        from
            terms
        where
            lang_id = (select lang_id from docs where doc_id = d.doc_id)
            and user_id = (select user_id from docs where doc_id = d.doc_id)
    )
group by
    value;

-- name: GetTerm :one
select
    term_id,
    value,
    translation
from 
    terms
where 
    value = @value 
    and lang_id = (select lang_id from docs d where d.doc_id = @doc_id) 
    and user_id = (select user_id from docs d where d.doc_id = @doc_id);

-- name: UpdateTerm :exec
update 
    terms
set 
    translation = case when @translation = '' then translation else @translation end, 
    term_level_id = case when @level_id = 0 then term_level_id else @level_id end
where
    term_id = @term_id;

-- name: UpdateDocStats :exec
update
    docs
set
    term_count = @term_count,
    sentence_count = @sentence_count
where
    doc_id = @doc_id;
    
