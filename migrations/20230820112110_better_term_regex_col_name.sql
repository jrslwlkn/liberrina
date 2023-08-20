-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table langs rename column trim_pattern to chars_pattern;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
