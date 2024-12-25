-- +goose Up
-- +goose StatementBegin
alter table docs drop column body;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
