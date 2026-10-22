-- +goose Up
-- +goose StatementBegin
CREATE TABLE links (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS links;
-- +goose StatementEnd
