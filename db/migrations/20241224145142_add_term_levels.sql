-- +goose Up
-- +goose StatementBegin
insert into
    term_levels(title)
values
    ('New'),
    ('Familiar'),
    ('Learned'),
    ('Known');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
truncate table term_levels;

-- +goose StatementEnd
