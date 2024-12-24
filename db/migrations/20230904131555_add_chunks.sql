-- +goose Up
-- +goose StatementBegin
create table chunks (
		doc_id integer not null,
		position integer not null,
		value string collate nocase not null,
		suffix string not null,

		PRIMARY KEY (doc_id, position)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists chunks;
-- +goose StatementEnd
