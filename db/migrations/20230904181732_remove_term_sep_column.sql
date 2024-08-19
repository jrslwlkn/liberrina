-- +goose Up
-- +goose StatementBegin
alter table langs drop term_sep;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table langs add column term_sep string;
-- +goose StatementEnd
