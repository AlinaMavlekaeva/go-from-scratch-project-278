-- +goose Up
-- +goose StatementBegin
ALTER TABLE links DROP CONSTRAINT links_short_url_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links
ADD CONSTRAINT links_short_url_key UNIQUE (short_url);
-- +goose StatementEnd
