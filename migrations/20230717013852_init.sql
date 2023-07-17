-- +goose Up
-- +goose StatementBegin

create table users (
	user_id integer not null primary key autoincrement,
	name text not null unique,
	created_at datetime not null
);

create table langs_dim (
    id text not null primary key,
    name text not null
);

create table langs (
	lang_id integer not null primary key autoincrement,
	name text unique not null,
    from_id text not null references langs_dim(id),
    to_id text not null references langs_dim(id),
    quick_lookup_uri text not null,
    lookup_uri_1 text,
    lookup_uri_2 text,
	word_sep text not null,
	trim_pattern text not null,
	sentence_sep text not null,
	user_id integer references users(user_id) not null,
	added_at datetime
);

create table docs (
	doc_id integer not null primary key autoincrement,
	title text not null,
	author text,
	body text not null,
	notes text,
	lang_id integer not null references langs(lang_id),
	user_id integer not null references users(user_id), 
	added_at datetime not null,
	word_count integer not null,
	sentence_count integer not null,
	words_new integer not null
);
create index idx_docs_user_id on docs(user_id);
create index idx_docs_lang_id on docs(lang_id);
create index idx_docs_lowercase_title on docs(lower(title));

create table word_levels (
	word_level_id integer not null primary key autoincrement,
	title text not null unique
);

create table words (
	word_id integer not null primary key autoincrement,
	value text not null,  
	translation text not null,
	word_level_id integer not null references word_levels(word_level_id),
	lang_id integer not null references langs(lang_id),
	user_id integer not null references users(user_id),  
	added_at datetime not null
);
create index idx_words_user_id on words(user_id);
create index idx_words_lang_id on words(lang_id);
create index idx_words_lowercase_value on words(lower(value));
	
create table sentences (
	sentence_id integer not null primary key autoincrement,
	body text not null,
	doc_id integer not null references docs(doc_id),
	word_count integer not null,
	words_new integer not null
);
create index idx_sentences_doc_id on sentences(doc_id);

create table word_to_sentence (
	word_id integer not null references words(word_id),
	sentence_id integer not null references sentences(sentence_id)
);
create index idx_word_to_sentence_word_id on word_to_sentence(word_id, sentence_id);
create index idx_word_to_sentence_sentence_id on word_to_sentence(sentence_id, word_id);

create table words_log (
	added_at timestamp not null,
	user_id integer not null references users(user_id),  
	word_id integer not null references words(word_id),  
	word_level_id integer not null references word_levels(word_level_id)
);
create index idx_words_log_user_id on words_log(user_id, word_id, word_level_id, added_at);

create table docs_log (
	opened_at timestamp not null,
	user_id integer not null references users(user_id),  
	doc_id integer not null references docs(doc_id)
);
create index idx_docs_log_user_id on docs_log(user_id, doc_id);
create index idx_docs_log_doc_id on docs_log(doc_id, user_id);

create table tags (
	tag_id integer not null primary key autoincrement,
	name text not null unique,
	added_at timestamp not null,
	user_id integer not null references users(user_id) 
);
create index idx_tags_user_id on tags(user_id);

create table word_to_tag (
	word_id integer not null references words(word_id),
	tag_id integer not null references tags(tag_id)
);
create index idx_word_to_tag_word_id on word_to_tag(word_id);
create index idx_word_to_tag_tag_id on word_to_tag(tag_id);

create table sentence_to_tag (
	sentence_id integer not null primary key autoincrement,
	tag_id integer not null references tags(tag_id)
);
create index idx_sentence_to_tag_word_id on sentence_to_tag(sentence_id);
create index idx_sentence_to_tag_tag_id on sentence_to_tag(tag_id);

create table doc_to_tag (
	doc_id integer not null primary key autoincrement,
	tag_id integer not null references tags(tag_id)  
);
create index idx_doc_to_tag_word_id on doc_to_tag(doc_id);
create index idx_doc_to_tag_tag_id on doc_to_tag(tag_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if exists words;
drop table if exists word_levels;
drop table if exists sentences;
drop table if exists tags;
drop table if exists users;
drop table if exists langs;
drop table if exists langs_dim;
drop table if exists docs;
drop table if exists doc_to_tag;
drop table if exists sentence_to_tag;
drop table if exists word_to_sentence;
drop table if exists word_to_tag;
drop table if exists docs_log;
drop table if exists words_log;

-- +goose StatementEnd
