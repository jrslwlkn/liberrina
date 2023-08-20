-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
PRAGMA foreign_keys=OFF;
drop index idx_terms_user_id;
drop index idx_terms_log_user_id;
drop index idx_tags_user_id;
drop index idx_docs_user_id;
drop index idx_docs_log_user_id;
drop index idx_docs_log_doc_id;
alter table docs_log drop column user_id;
alter table terms_log drop column user_id;
alter table terms drop column user_id;
alter table tags drop column user_id;
alter table docs drop column user_id;
alter table langs drop column user_id;
drop table users;
PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
