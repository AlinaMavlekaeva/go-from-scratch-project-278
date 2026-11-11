-- +goose Up
-- +goose StatementBegin
ALTER TABLE links ADD CONSTRAINT links_short_name_key UNIQUE (short_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links DROP CONSTRAINT links_short_name_key;
-- +goose StatementEnd
