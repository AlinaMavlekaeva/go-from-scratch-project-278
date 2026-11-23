-- +goose Up
-- +goose StatementBegin
CREATE TABLE link_visits(
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    ip TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    referer TEXT,
    status INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS link_visits;
-- +goose StatementEnd
